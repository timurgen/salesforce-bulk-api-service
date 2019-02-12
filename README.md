# salesforce-bulk-api-service
Simple service for quering Salesforce data via Bulk API

## Sesam system set up 
```json
{
  "_id": "<system id>",
  "type": "system:microservice",
  "docker": {
    "environment": {
      "SALESFORCE_PASSWORD": "$SECRET(salesforce-password)",
      "SALESFORCE_USERNAME": "$ENV(salesforce-user)",
      "SALESFORCE_USER_TOKEN": "$SECRET(salesforce-security_token)",
      "SANDBOX": "any non empty value if you wish to use Salesforce sandbox instead of prod env",
      "DEBUG": "any non empty value if need to use in debug mode"
    },
    "image": "ohuenno/salesforceclient:latest",
    "memory": 2048,
    "port": 8080
  }
}
```

## Sesam pipe set up

```json
{
  "_id": "<pipe id>",
  "type": "pipe",
  "source": {
    "type": "conditional",
    "alternatives": {
      "prod": {
        "type": "json",
        "system": "salesforce",
        "supports_since": true,
        "url": "Contact"
      },
      "test": {
        "type": "json",
        "system": "salesforce",
        "supports_since": true,
        "url": "Contact"
      }
    },
    "condition":  "prod"
  },
  "transform": {
    "type": "dtl",
    "rules": {
      "default": [
        ["copy", "*"],
        ["add", "_id", "_S.Id"]
      ]
    }
  },
  "pump": {
    "cron_expression": "0 0 * * ?"
  }
}

```
