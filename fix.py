import re

with open("ui/src/components/credentials/credential-form.tsx", "r") as f:
    content = f.read()

# Replace auth in handleTest
search = """        const auth: Authentication = {}
        if (values.authType === "api_key") {
            auth.apiKey = {
                paramName: values.apiKeyParamName || "X-API-Key",
                in: parseInt(values.apiKeyLocation || "0") as APIKeyAuth_Location,
                value: { plainText: values.apiKeyValue || "", validationRegex: "" },
                verificationValue: ""
            }
        } else if (values.authType === "bearer_token") {
             auth.bearerToken = { token: { plainText: values.bearerToken || "", validationRegex: "" } }
        } else if (values.authType === "basic_auth") {
            auth.basicAuth = { username: values.basicUsername || "", password: { plainText: values.basicPassword || "", validationRegex: "" }, passwordHash: "" }
        } else if (values.authType === "oauth2") {"""

replace = """        const auth: any = {}
        if (values.authType === "api_key") {
            auth.api_key = {
                param_name: values.apiKeyParamName || "X-API-Key",
                in: parseInt(values.apiKeyLocation || "0"),
                value: { plain_text: values.apiKeyValue || "", validation_regex: "" },
                verification_value: ""
            }
        } else if (values.authType === "bearer_token") {
            auth.bearer_token = {
                token: { plain_text: values.bearerToken || "", validation_regex: "" }
            }
        } else if (values.authType === "basic_auth") {
            auth.basic_auth = {
                username: values.basicUsername || "",
                password: { plain_text: values.basicPassword || "", validation_regex: "" },
                password_hash: ""
            }
        } else if (values.authType === "oauth2") {"""

content = content.replace(search, replace)

search2 = """                 auth.oauth2 = {
                     clientId: { plainText: values.oauthClientId || "", validationRegex: "" },
                     clientSecret: { plainText: values.oauthClientSecret || "", validationRegex: "" },
                     authorizationUrl: values.oauthAuthUrl || "",
                     tokenUrl: values.oauthTokenUrl || "",
                     scopes: values.oauthScopes || "",
                     issuerUrl: "",
                     audience: "",
                 }
             } else {
                 auth.oauth2 = {
                     clientId: { plainText: values.oauthClientId || "", validationRegex: "" },
                     clientSecret: { plainText: values.oauthClientSecret || "", validationRegex: "" },
                     authorizationUrl: values.oauthAuthUrl || "",
                     tokenUrl: values.oauthTokenUrl || "",
                     scopes: values.oauthScopes || "",
                     issuerUrl: "",
                     audience: "",
                 }"""

replace2 = """                 auth.oauth2 = {
                     client_id: { plain_text: values.oauthClientId || "", validation_regex: "" },
                     client_secret: { plain_text: values.oauthClientSecret || "", validation_regex: "" },
                     authorization_url: values.oauthAuthUrl || "",
                     token_url: values.oauthTokenUrl || "",
                     scopes: values.oauthScopes || "",
                     issuer_url: "",
                     audience: "",
                 }
             } else {
                 auth.oauth2 = {
                     client_id: { plain_text: values.oauthClientId || "", validation_regex: "" },
                     client_secret: { plain_text: values.oauthClientSecret || "", validation_regex: "" },
                     authorization_url: values.oauthAuthUrl || "",
                     token_url: values.oauthTokenUrl || "",
                     scopes: values.oauthScopes || "",
                     issuer_url: "",
                     audience: "",
                 }"""

content = content.replace(search2, replace2)

with open("ui/src/components/credentials/credential-form.tsx", "w") as f:
    f.write(content)
