{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Image Tagging Schema",
  "description": "Schema for tagging and classifying images in a web-based portal",
  "type": "object",
  "required": ["imageId", "classifications"],
  "properties": {
    "imageId": {
      "type": "string",
      "description": "Unique identifier for the image",
      "pattern": "^[a-zA-Z0-9_-]+$"
    },
    "classifications": {
      "type": "object",
      "description": "Classification tags for the image",
      "required": ["people", "dateRange", "occasion"],
      "properties": {
        "people": {
          "type": "array",
          "description": "List of people identified in the image",
          "items": {
            "type": "object",
            "required": ["name"],
            "properties": {
              "name": {
                "type": "string",
                "description": "Full name of the person",
                "minLength": 1
              }
            }
          }
        },
        "dateRange": {
          "type": "object",
          "description": "Date or date range when the image was taken",
          "required": ["type"],
          "properties": {
            "type": {
              "type": "string",
              "enum": ["exact", "range", "approximate"],
              "description": "Type of date classification"
            },
            "exactDate": {
              "type": "string",
              "format": "date",
              "description": "Exact date when type is 'exact' (YYYY-MM-DD)"
            },
            "startDate": {
              "type": "string",
              "format": "date",
              "description": "Start date when type is 'range'"
            },
            "endDate": {
              "type": "string",
              "format": "date",
              "description": "End date when type is 'range'"
            },
            "approximateDate": {
              "type": "object",
              "description": "Approximate date information",
              "properties": {
                "year": {
                  "type": "integer"
                },
                "month": {
                  "type": "integer",
                  "minimum": 1,
                  "maximum": 12
                }
              }
            }
          }
        },
        "occasion": {
          "type": "object",
          "description": "Event or occasion associated with the image",
          "required": ["category"],
          "properties": {
            "category": {
              "type": "string",
              "enum": [
                "birthday",
                "wedding",
                "graduation",
                "holiday",
                "vacation",
                "work_event",
                "party",
                "family_gathering",
                "sports_event",
                "concert",
                "conference",
                "ceremony",
                "casual",
                "other"
              ],
              "description": "Primary occasion category"
            },
            "eventName": {
              "type": "string",
              "description": "Specific name of the event"
            }
          }
        }
      }
    }
  }
}
