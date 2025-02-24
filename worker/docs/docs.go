// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "consumes": [
        "application/json"
    ],
    "produces": [
        "application/json"
    ],
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Petrov Sergey",
            "email": "s.petrov1@g.nsu.ru"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "https://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/manager/health/liveness": {
            "get": {
                "description": "Request for getting health liveness.",
                "tags": [
                    "Health API"
                ],
                "summary": "Health liveness",
                "operationId": "healthLiveness",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/api/manager/health/readiness": {
            "get": {
                "description": "Request for getting health readiness. In response will be status of all check (database, cache, message queue).",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Health API"
                ],
                "summary": "Health readiness",
                "operationId": "healthReadiness",
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "503": {
                        "description": "Service Unavailable",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorOutput"
                        }
                    }
                }
            }
        },
        "/api/manager/swagger/api-docs.json": {
            "get": {
                "description": "Request for getting swagger specification in JSON",
                "produces": [
                    "application/json; charset=utf-8"
                ],
                "tags": [
                    "Swagger API"
                ],
                "summary": "Swagger JSON",
                "operationId": "SwaggerJSON",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/manager/swagger/index.html": {
            "get": {
                "description": "Request for getting swagger UI",
                "produces": [
                    "text/html; charset=utf-8"
                ],
                "tags": [
                    "Swagger API"
                ],
                "summary": "Swagger UI",
                "operationId": "SwaggerUI",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/internal/api/worker/hash/crack/task": {
            "post": {
                "description": "Request for executing hash crack task.",
                "consumes": [
                    "application/xml"
                ],
                "produces": [
                    "application/xml"
                ],
                "tags": [
                    "Hash Crack Task API"
                ],
                "summary": "Hash crack task",
                "operationId": "hashCrackTask",
                "parameters": [
                    {
                        "description": "Hash crack task input",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.HashCrackTaskInput"
                        }
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Accepted"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorOutput"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/model.ErrorOutput"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.ErrorOutput": {
            "type": "object",
            "required": [
                "message",
                "path",
                "status",
                "timestamp"
            ],
            "properties": {
                "message": {
                    "type": "string"
                },
                "path": {
                    "type": "string",
                    "format": "url_path",
                    "example": "/api/v0/example"
                },
                "status": {
                    "type": "integer",
                    "maximum": 599,
                    "minimum": 400
                },
                "timestamp": {
                    "type": "string"
                }
            }
        },
        "model.HashCrackTaskInput": {
            "type": "object",
            "required": [
                "alphabet",
                "hash",
                "requestID"
            ],
            "properties": {
                "alphabet": {
                    "type": "object",
                    "properties": {
                        "symbols": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    }
                },
                "hash": {
                    "type": "string"
                },
                "maxLength": {
                    "type": "integer",
                    "maximum": 6,
                    "minimum": 0
                },
                "partCount": {
                    "type": "integer"
                },
                "partNumber": {
                    "type": "integer"
                },
                "requestID": {
                    "type": "string"
                }
            }
        }
    },
    "tags": [
        {
            "description": "API for cracking hashes and sending results",
            "name": "Hash Crack Task API"
        },
        {
            "description": "API for health checks",
            "name": "Health API"
        },
        {
            "description": "API for getting swagger specification",
            "name": "Swagger API"
        }
    ],
    "externalDocs": {
        "description": "OpenAPI",
        "url": "https://swagger.io/resources/open-api/"
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.0.0",
	Host:             "localhost:8080",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "Crack Hash Worker API",
	Description:      "API for Crack Hash Worker",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
