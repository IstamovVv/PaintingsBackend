package endpoint

import (
	"encoding/base64"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"paint-backend/internal/repo"
	"paint-backend/internal/s3"
	"paint-backend/internal/util/cast"
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
	"/api/v1/products/all": {
		handler: func(ctx *fasthttp.RequestCtx, h *HttpHandler) {
			if cast.ByteArrayToString(ctx.Method()) == fasthttp.MethodGet {
				h.getAllProducts(ctx)
			} else {
				ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			}
		},
	},

	"/api/v1/products": {
		handler: func(ctx *fasthttp.RequestCtx, h *HttpHandler) {
			switch cast.ByteArrayToString(ctx.Method()) {
			case fasthttp.MethodPost:
				{
					h.insertProduct(ctx)
				}
			case fasthttp.MethodDelete:
				{
					h.deleteProduct(ctx)
				}
			default:
				{
					ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
				}
			}
		},
	},

	"/api/v1/images": {
		handler: func(ctx *fasthttp.RequestCtx, h *HttpHandler) {
			switch cast.ByteArrayToString(ctx.Method()) {
			case fasthttp.MethodPost:
				{
					h.insertImage(ctx)
				}
			case fasthttp.MethodDelete:
				{
					h.deleteImage(ctx)
				}
			default:
				{
					ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
				}
			}
		},
	},
}

type HttpHandler struct {
	storage       *s3.Storage
	productsTable *repo.ProductsTable
}

func NewHttpHandler(storage *s3.Storage, table *repo.ProductsTable) *HttpHandler {
	return &HttpHandler{
		storage:       storage,
		productsTable: table,
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

	if r, ok := routingMap[cast.ByteArrayToString(ctx.Path())]; ok {
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

	products, err := h.productsTable.GetAllProducts(offset, limit)
	writeObject(ctx, products, fasthttp.StatusOK)
}

func (h *HttpHandler) insertProduct(ctx *fasthttp.RequestCtx) {
	var product repo.Product
	err := json.Unmarshal(ctx.PostBody(), &product)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	if len(product.Images) > 10 {
		writeError(ctx, "Too many images", fasthttp.StatusBadRequest)
		return
	}

	err = h.productsTable.Insert(product)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
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
	images, err := h.storage.GetAllImages()
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	writeObject(ctx, images, fasthttp.StatusOK)
}

func (h *HttpHandler) insertImage(ctx *fasthttp.RequestCtx) {
	nameBytes := ctx.QueryArgs().Peek("name")
	if len(nameBytes) == 0 {
		writeError(ctx, "empty name", fasthttp.StatusBadRequest)
		return
	}

	name := cast.ByteArrayToString(nameBytes)

	var image string
	err := json.Unmarshal(ctx.PostBody(), &image)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusBadRequest)
		return
	}

	decodedImage, err := base64.StdEncoding.DecodeString(image)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	err = h.storage.InsertImage(name, decodedImage)
	if err != nil {
		writeError(ctx, err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
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
