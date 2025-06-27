package swagger

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/ptrvsrg/crack-hash/commonlib/http/handler"
	"github.com/ptrvsrg/crack-hash/manager/docs"
)

type hdlr struct {
	logger zerolog.Logger
}

func NewHandler(logger zerolog.Logger) handler.Handler {
	return &hdlr{
		logger: logger.With().Str("handler", "swagger").Logger(),
	}
}

func (h *hdlr) RegisterRoutes(router *gin.Engine) {
	h.logger.Debug().Msg("register swagger routes")

	swaggerRouter := router.Group("/swagger")
	{
		swaggerRouter.GET("/api-docs.json", h.getAPIDocJSON)
		swaggerRouter.GET("/index.html", h.getUI)
	}
}

// getUI godoc
//
//	@Id				SwaggerUI
//	@Summary		Swagger UI
//	@Description	Request for getting swagger UI
//	@Tags			Swagger API
//	@Produce		text/html; charset=utf-8
//	@Success		200	{object}	string
//	@Router			/swagger/index.html [get]
func (h *hdlr) getUI(ctx *gin.Context) {
	h.logger.Debug().Msg("get swagger UI")

	swaggerJSON := docs.SwaggerInfo.ReadDoc()
	swaggerUI := renderUITemplate(swaggerJSON)

	ctx.Data(http.StatusOK, "text/html; charset=utf-8", swaggerUI)
}

// getAPIDocJSON godoc
//
//	@Id				SwaggerJSON
//	@Summary		Swagger JSON
//	@Description	Request for getting swagger specification in JSON
//	@Tags			Swagger API
//	@Produce		application/json; charset=utf-8
//	@Success		200	{object}	string
//	@Router			/swagger/api-docs.json [get]
func (h *hdlr) getAPIDocJSON(ctx *gin.Context) {
	h.logger.Debug().Msg("get swagger JSON")

	swaggerJSON := docs.SwaggerInfo.ReadDoc()

	ctx.Data(http.StatusOK, "application/json; charset=utf-8", []byte(swaggerJSON))
}

func renderUITemplate(swaggerJSON string) []byte {
	template := `<!DOCTYPE html>
<html>
<head>
   <meta charset="UTF-8">
   <link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.19.5/swagger-ui.css" >
   <style>
       .topbar {
           display: none;
       }
   </style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.19.5/swagger-ui-bundle.js"> </script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.19.5/swagger-ui-standalone-preset.js"> </script>
<script>
	const spec = %s;
	window.onload = function() {
		// Build a system
		const ui = SwaggerUIBundle({
			dom_id: '#swagger-ui',
			deepLinking: true,
			spec: spec,
			presets: [
				SwaggerUIBundle.presets.apis,
				SwaggerUIStandalonePreset
			],
			plugins: [
				SwaggerUIBundle.plugins.DownloadURL
			],
			layout: "BaseLayout",
		})
		window.ui = ui
	}
</script>
</body>
</html>`

	return []byte(fmt.Sprintf(template, swaggerJSON))
}
