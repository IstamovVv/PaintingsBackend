package endpoint

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"net/http"
	"paint-backend/internal/repo"
	"paint-backend/internal/s3"
	"paint-backend/internal/util/cast"
	"strconv"
	"strings"
	"sync"
)

const (
	maxImageSizeInBytes = 1024 * 1024 * 10
)

var (
	mutex sync.Mutex
)

func init() {
	for path, info := range routingMap {
		info.path = path
		routingMap[path] = info
	}
}

type route struct {
	handler func(ctx *fasthttp.RequestCtx, h *HttpHandler)
	path    string
}

var routingMap = map[string]route{
	"/api/v1/products": {
		handler: func(ctx *fasthttp.RequestCtx, h *HttpHandler) {
			switch cast.ByteArrayToString(ctx.Method()) {
			case fasthttp.MethodGet:
				h.getAllProducts(ctx)
			case fasthttp.MethodPut:
				h.insertProduct(ctx)
			case fasthttp.MethodDelete:
				h.deleteProduct(ctx)
			default:
				ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			}
		},
	},

	"/api/v1/images": {
		handler: func(ctx *fasthttp.RequestCtx, h *HttpHandler) {
			switch cast.ByteArrayToString(ctx.Method()) {
			case fasthttp.MethodGet:
				h.getAllImages(ctx)
			case fasthttp.MethodPost:
				h.insertImage(ctx)
			case fasthttp.MethodDelete:
				h.deleteImage(ctx)
			default:
				ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			}
		},
	},

	"/api/v1/images/folders": {
		handler: func(ctx *fasthttp.RequestCtx, h *HttpHandler) {
			switch cast.ByteArrayToString(ctx.Method()) {
			case fasthttp.MethodGet:
				h.getImagesFolders(ctx)
			default:
				ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			}
		},
	},

	"/api/v1/brands": {
		handler: func(ctx *fasthttp.RequestCtx, h *HttpHandler) {
			switch cast.ByteArrayToString(ctx.Method()) {
			case fasthttp.MethodGet:
				h.getAllBrands(ctx)
			case fasthttp.MethodPost:
				h.insertBrand(ctx)
			case fasthttp.MethodDelete:
				h.deleteBrand(ctx)
			default:
				ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			}
		},
	},

	"/api/v1/brands-by-subject": {
		handler: func(ctx *fasthttp.RequestCtx, h *HttpHandler) {
			switch cast.ByteArrayToString(ctx.Method()) {
			case fasthttp.MethodGet:
				h.getBrandsBySubject(ctx)
			}
		},
	},

	"/api/v1/subjects": {
		handler: func(ctx *fasthttp.RequestCtx, h *HttpHandler) {
			switch cast.ByteArrayToString(ctx.Method()) {
			case fasthttp.MethodGet:
				h.getAllSubjects(ctx)
			case fasthttp.MethodPost:
				h.insertSubject(ctx)
			case fasthttp.MethodDelete:
				h.deleteSubject(ctx)
			default:
				ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			}
		},
	},
}

type HttpHandler struct {
	storage           *s3.Storage
	productsTable     *repo.ProductsTable
	subjectsTable     *repo.SubjectsTable
	brandsTable       *repo.BrandsTable
	subjectBrandTable *repo.SubjectBrandTable
}

func NewHttpHandler(storage *s3.Storage, productsTable *repo.ProductsTable, subjectsTable *repo.SubjectsTable, brandsTable *repo.BrandsTable, subjectBrandTable *repo.SubjectBrandTable) *HttpHandler {
	return &HttpHandler{
		storage:           storage,
		productsTable:     productsTable,
		subjectsTable:     subjectsTable,
		brandsTable:       brandsTable,
		subjectBrandTable: subjectBrandTable,
	}
}

func (h *HttpHandler) Handle(ctx *fasthttp.RequestCtx) {
	defer func() {
		err := recover()
		if err != nil {
			logrus.Error(err)
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		}
	}()

	// TODO replace mutex by conn pool
	mutex.Lock()
	defer mutex.Unlock()

	if r, ok := routingMap[cast.ByteArrayToString(ctx.Path())]; ok {
		addCorsHeaders(ctx)

		if cast.ByteArrayToString(ctx.Method()) == fasthttp.MethodOptions {
			ctx.SetStatusCode(fasthttp.StatusOK)
			return
		}

		r.handler(ctx, h)
	} else {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	}
}

func (h *HttpHandler) getAllProducts(ctx *fasthttp.RequestCtx) {
	offset, err := ctx.QueryArgs().GetUint("offset")
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	limit, err := ctx.QueryArgs().GetUint("limit")
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	var searchOptions repo.SearchProductsOptions

	subjectBytes := ctx.QueryArgs().Peek("subject")
	if len(subjectBytes) != 0 {
		searchOptions.SubjectFilter = true
		searchOptions.Subject = cast.ByteArrayToString(subjectBytes)
	}

	brandBytes := ctx.QueryArgs().Peek("brand")
	if len(brandBytes) != 0 {
		searchOptions.BrandFilter = true
		searchOptions.Brand = cast.ByteArrayToString(brandBytes)
	}

	products, err := h.productsTable.GetAllProducts(offset, limit, searchOptions)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	if products == nil {
		products = []repo.Product{}
	}

	writeObject(ctx, products, fasthttp.StatusOK)
}

func (h *HttpHandler) insertProduct(ctx *fasthttp.RequestCtx) {
	editFlagBytes := ctx.QueryArgs().Peek("edit")
	if len(editFlagBytes) == 0 {
		writeError(ctx, "empty edit flag", fasthttp.StatusBadRequest)
		return
	}

	editFlag, err := strconv.ParseBool(cast.ByteArrayToString(editFlagBytes))
	if err != nil {
		writeError(ctx, "failed to parse edit flag: "+err.Error(), fasthttp.StatusBadRequest)
		return
	}

	var product repo.Product
	err = json.Unmarshal(ctx.PostBody(), &product)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	if len(product.Images) > 10 {
		writeError(ctx, "Too many images", fasthttp.StatusBadRequest)
		return
	}

	err = h.productsTable.Insert(product, editFlag)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	var relatedBrands []uint
	var brandFound bool

	relatedBrands, err = h.subjectBrandTable.GetBrandIdsBySubjectId(product.SubjectId)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	for _, brandId := range relatedBrands {
		if product.BrandId == brandId {
			brandFound = true
			break
		}
	}

	if !brandFound {
		err = h.subjectBrandTable.Insert(repo.SubjectBrand{
			SubjectId: product.SubjectId,
			BrandId:   product.BrandId,
		})

		if err != nil {
			writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
			return
		}
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (h *HttpHandler) deleteProduct(ctx *fasthttp.RequestCtx) {
	id, err := ctx.QueryArgs().GetUint("id")
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	err = h.productsTable.Delete(uint(id))
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (h *HttpHandler) getAllImages(ctx *fasthttp.RequestCtx) {
	path := cast.ByteArrayToString(ctx.QueryArgs().Peek("path"))
	images, err := h.storage.GetImages(strings.Join(strings.Split(path, ","), "/"))
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	writeObject(ctx, images, fasthttp.StatusOK)
}

func (h *HttpHandler) getImagesFolders(ctx *fasthttp.RequestCtx) {
	folders, err := h.storage.GetImagesFolders()
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	writeObject(ctx, folders, fasthttp.StatusOK)
}

func (h *HttpHandler) insertImage(ctx *fasthttp.RequestCtx) {
	nameBytes := ctx.QueryArgs().Peek("name")
	if len(nameBytes) == 0 {
		writeError(ctx, "empty name", fasthttp.StatusBadRequest)
		return
	}

	name := cast.ByteArrayToString(nameBytes)
	body := ctx.PostBody()

	if len(body) > maxImageSizeInBytes {
		writeError(ctx, "Too big image", fasthttp.StatusBadRequest)
		return
	}

	mimeType := http.DetectContentType(body)
	if mimeType != "image/png" && mimeType != "image/jpeg" {
		writeError(ctx, fmt.Sprintf("Invalid image type: %s. Allowed only jpeg and png", mimeType), fasthttp.StatusBadRequest)
		return
	}

	err := h.storage.InsertImage(name, mimeType, body)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	writeObject(ctx, name, fasthttp.StatusOK)
}

func (h *HttpHandler) deleteImage(ctx *fasthttp.RequestCtx) {
	nameBytes := ctx.QueryArgs().Peek("name")
	if len(nameBytes) == 0 {
		writeError(ctx, "empty name", fasthttp.StatusBadRequest)
		return
	}

	name := cast.ByteArrayToString(nameBytes)

	err := h.storage.DeleteImage(name)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (h *HttpHandler) getAllBrands(ctx *fasthttp.RequestCtx) {
	brands, err := h.brandsTable.GetAll()
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	if brands == nil {
		brands = []repo.Brand{}
	}

	writeObject(ctx, brands, fasthttp.StatusOK)
}

func (h *HttpHandler) getBrandsBySubject(ctx *fasthttp.RequestCtx) {
	id, err := ctx.QueryArgs().GetUint("subject_id")
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	brands, err := h.subjectBrandTable.GetBrandIdsBySubjectId(uint(id))
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	if brands == nil {
		brands = []uint{}
	}

	writeObject(ctx, brands, fasthttp.StatusOK)
}

func (h *HttpHandler) insertBrand(ctx *fasthttp.RequestCtx) {
	var brand repo.Brand
	err := json.Unmarshal(ctx.PostBody(), &brand)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	err = h.brandsTable.Insert(brand)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (h *HttpHandler) deleteBrand(ctx *fasthttp.RequestCtx) {
	id, err := ctx.QueryArgs().GetUint("id")
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	err = h.brandsTable.Delete(uint(id))
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (h *HttpHandler) getAllSubjects(ctx *fasthttp.RequestCtx) {
	subjects, err := h.subjectsTable.GetAll()
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	if subjects == nil {
		subjects = []repo.Subject{}
	}

	writeObject(ctx, subjects, fasthttp.StatusOK)
}

func (h *HttpHandler) insertSubject(ctx *fasthttp.RequestCtx) {
	var subject repo.Subject
	err := json.Unmarshal(ctx.PostBody(), &subject)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	err = h.subjectsTable.Insert(subject)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (h *HttpHandler) deleteSubject(ctx *fasthttp.RequestCtx) {
	id, err := ctx.QueryArgs().GetUint("id")
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	err = h.subjectsTable.Delete(uint(id))
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeObject(ctx *fasthttp.RequestCtx, obj any, status int) {
	row, err := json.Marshal(obj)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(status)
	ctx.Response.Header.Set(fasthttp.HeaderContentType, "application/json")
	_, _ = ctx.Write(row)
}

func writeError(ctx *fasthttp.RequestCtx, message string, status int) {
	response := errorResponse{Error: message}
	row, err := json.Marshal(&response)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(status)
	ctx.Response.Header.Set(fasthttp.HeaderContentType, "application/json")
	_, _ = ctx.Write(row)
}

func addCorsHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH")
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "*")
}
