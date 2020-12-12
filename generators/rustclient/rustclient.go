package rustclient

import (
	"fmt"
	"io"

	"github.com/apex/rpc/internal/format"
	"github.com/apex/rpc/schema"
)

var error_handling = `// Error is an error returned by the client.
#[derive(Serialize, Deserialize, Debug, Clone, Default)]
pub struct ClientError {
    status: String,
    status_code: u16,
    #[serde(rename = "type")]
    err_type: Option<String>,
    message: Option<String>,
}

impl From<serde_json::error::Error> for ClientError {
    fn from(err: serde_json::error::Error) -> ClientError {
        ClientError {
            status: "Internal Server Error".into(),
            status_code: 500,
            err_type: Some("json".into()),
            message: Some(err.to_string()),
        }
    }
}

impl From<reqwest::Error> for ClientError {
    fn from(err: reqwest::Error) -> ClientError {
        ClientError {
            status: "Internal Server Error".into(),
            status_code: 500,
            err_type: Some("reqwest".into()),
            message: Some(err.to_string()),
        }
    }
}
`

var call = `// call implementation.
    async fn call(
        &self,
        method: &str,
        input: Option<Vec<u8>>,
    ) -> Result<bytes::Bytes, ClientError> {
        let uri = format!("{}/{}", self.endpoint, method);

        let mut builder = self
            .client
            .post(&uri)
            .header("Content-Type", "application/json");

        if let Some(data) = input {
            builder = builder.body(data);
        }

        if self.auth_token.is_some() {
            builder = builder.header("Authorization", format!("Bearer {:?}", &self.auth_token));
        }

        let resp = builder.send().await?;

        let status_code = resp.status();
        if status_code.as_u16() > 300 {
            let mut e = ClientError {
                ..Default::default()
            };

            if let Some(content_type) = resp.headers().get("Content-Type") {
                if content_type == "application/json" {
                    let body = resp.bytes().await?;
                    e = serde_json::from_slice::<ClientError>(&body)?;
                }
            }

            e.status_code = status_code.as_u16();
            e.status = status_code.canonical_reason().unwrap_or_default().into();

            return Err(e);
        }

        let body = resp.bytes().await?;
        return Ok(body);
    }
`

// Generate writes the Go client implementations to w.
func Generate(w io.Writer, s *schema.Schema) error {
	out := fmt.Fprintf

	out(w, "// Client is the API client.\n")
	out(w, "#[derive(Debug, Clone)]\n")
	out(w, "pub struct Client {\n")
	out(w, "  client: reqwest::Client,\n")
	out(w, "  endpoint: String,\n")
	out(w, "  auth_token: Option<String>,\n")
	out(w, "}\n\n")

	out(w, "impl Client {\n\n")

	out(w, "  pub fn new(client: reqwest::Client, endpoint: &str, auth_token: Option<String>) -> Client{\n")
	out(w, "    Client {\n")
	out(w, "      client: client,\n")
	out(w, "      endpoint: endpoint.to_string(), \n")
	out(w, "      auth_token: auth_token\n")
	out(w, "    }\n")
	out(w, "  }\n\n")

	for _, m := range s.Methods {
		name := format.GoName(m.Name)
		rname := format.RustName(m.Name)
		out(w, "  // %s\n", m.Description)
		out(w, "  pub async fn %s(&self", rname)

		// input arg
		if len(m.Inputs) > 0 {
			out(w, ", input: &%sInput", name)
		}
		out(w, ") ")

		// output arg
		if len(m.Outputs) > 0 {
			out(w, "-> Result<%sOutput, ClientError> {\n", name)
		} else {
			out(w, "-> Result<(), ClientError> {\n")
		}

		// Serialize any inputs
		if len(m.Inputs) > 0 {
			out(w, "    let json = serde_json::to_vec(input)?;\n")
		}

		if len(m.Outputs) > 0 {
			out(w, "    let res: bytes::Bytes = ")
		} else {
			out(w, "    ")
		}

		out(w, "self.call(\"%s\", ", m.Name)
		if len(m.Inputs) > 0 {
			out(w, "Some(json)")
		} else {
			out(w, "None")
		}

		out(w, ").await?;\n")
		if len(m.Outputs) > 0 {
			out(w, "    let output: %sOutput = serde_json::from_slice(&res)?;\n", name)
			out(w, "    return Ok(output)\n")
		} else {
			out(w, "    Ok(())\n")
		}

		// close
		out(w, "  }\n\n")
	}

	out(w, "\n%s\n", call)

	out(w, "}\n\n")

	out(w, "\n%s\n", error_handling)

	return nil
}
