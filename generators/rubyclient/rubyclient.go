package rubyclient

import (
	"fmt"
	"io"
	"strings"

	"github.com/apex/rpc/schema"
)

var module = `
# Do not edit, this file was generated by github.com/apex/rpc.

require 'net/http'
require 'net/https'
require 'json'

module %s
  class %s
    # Error is raised when an API call fails due to a 4xx or 5xx HTTP error.
    class Error < StandardError
      attr_reader :type
      attr_reader :message
      attr_reader :status

      def initialize(status, type = nil, message = nil)
        @status = status
        @type = type
        @message = message
      end

      def to_s
        if @type
          "#{@status} response: #{@type}: #{@message}"
        else
          "#{@status} response"
        end
      end
    end

    # Initialize the client with API endpoint URL and optional authentication token.
    def initialize(url, auth_token = nil)
      @url = url
      @auth_token = auth_token
    end
`

var call = `
    private
  
    # call an API method with optional input parameters.
    def call(method, params = nil)
      url = @url + "/" + method
      header = { "Content-Type" => "application/json" }
  
      if @auth_token
        header["Authorization"] = "Bearer #{@auth_token}"
      end
  
      res = Net::HTTP.post URI(url), params.to_json, header
      status = res.code.to_i
  
      if status >= 400
        begin
          body = JSON.parse(res.body)
        rescue
          raise Error.new(status)
        end
        raise Error.new(status, body["type"], body["message"])
      end
  
      res.body
    end
`

// Generate writes the Ruby client implementations to w.
func Generate(w io.Writer, s *schema.Schema, moduleName, className string) error {
	out := fmt.Fprintf

	out(w, module, moduleName, className)

	for _, m := range s.Methods {
		// comment
		out(w, "\n")
		out(w, "    # %s\n", capitalize(m.Description))
		if len(m.Inputs) > 0 {
			out(w, "    #\n")
			out(w, "    # @param [Hash] params the input for this method.\n")
			for _, f := range m.Inputs {
				out(w, "    # @param params [%s] :%s %s\n", rubyType(s, f), f.Name, capitalize(f.Description))
			}
		}

		// method
		out(w, "    def %s", m.Name)

		// input arg
		if len(m.Inputs) > 0 {
			out(w, "(params)")
		}
		out(w, "\n")

		// return
		if len(m.Inputs) > 0 {
			out(w, "      call %q, params\n", m.Name)
		} else {
			out(w, "      call %q\n", m.Name)
		}

		// close
		out(w, "    end\n")
	}

	out(w, "%s", call)
	out(w, "  end\n")
	out(w, "end\n")

	return nil
}

// capitalize returns a capitalized string.
func capitalize(s string) string {
	return strings.ToUpper(string(s[0])) + string(s[1:])
}

// rubyType returns a Ruby equivalent type for field f.
func rubyType(s *schema.Schema, f schema.Field) string {
	// TODO: handle reference types, not sure if makes sense
	// to generate classes for Ruby inputs or not
	switch f.Type.Type {
	case schema.String:
		return "String"
	case schema.Int, schema.Float:
		return "Number"
	case schema.Bool:
		return "Boolean"
	case schema.Timestamp:
		return "Date"
	case schema.Object:
		return "Hash"
	case schema.Array:
		return "Array"
	default:
		return ""
	}
}
