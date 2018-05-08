package imglib

var schema = string(`
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "AndroidX Config",
  "additionalProperties": false,
  "definitions": {
    "kernel": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "image": {"type": "string"},
        "cmdline": {"type": "string"}
      }
    },
    "trust": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "image": { "$ref": "#/definitions/strings" },
        "org": { "$ref": "#/definitions/strings" }
      }
    },
    "strings": {
        "type": "array",
        "items": {"type": "string"}
    }
  },
  "properties": {
    "kernel": { "$ref": "#/definitions/kernel" },
    "init": { "$ref": "#/definitions/strings" },
    "trust": { "$ref": "#/definitions/trust" }
  }
}
`)
