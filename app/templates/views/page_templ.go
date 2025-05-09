// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.793
package views

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"wonk/app/templates/components"
)

func Page() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<!doctype html><html><head><meta charset=\"UTF-8\"><title>Wonk</title><link rel=\"stylesheet\" href=\"static/css/output.css\"><script src=\"/static/script/htmx.min.js\"></script><script src=\"/static/script/hyperscript.min.js\"></script><script>\n\t\tdocument.addEventListener('DOMContentLoaded', (event) => {\n\t\t\tdocument.body.addEventListener('htmx:beforeSwap', function (evt) {\n\t\t\t\tif (evt.detail.xhr.status === 404) {\n\t\t\t\t\t// alert the user when a 404 occurs (maybe use a nicer mechanism than alert())\n\t\t\t\t\talert(\"Error: Could Not Find Resource\");\n\t\t\t\t} else if (evt.detail.xhr.status === 422) {\n\t\t\t\t\t// allow 422 responses to swap as we are using this as a signal that\n\t\t\t\t\t// a form was submitted with bad data and want to rerender with the\n\t\t\t\t\t// errors\n\t\t\t\t\t//\n\t\t\t\t\t// set isError to false to avoid error logging in console\n\t\t\t\t\tevt.detail.shouldSwap = true;\n\t\t\t\t\tevt.detail.isError = false;\n\t\t\t\t}\n\t\t\t});\n\t\t})\n\t</script></head><body class=\"overscroll-none light text-txt-primary bg-bg-main\"><div class=\"h-screen\"><div class=\"h-full flex flex-col divide-y-1 divide-brdr-main\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = components.Header().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"flex flex-row h-[94%] divide-x-1 divide-brdr-main\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = components.NavBar().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div id=\"content-div\" class=\"w-full p-4\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div></div></div></div></body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
