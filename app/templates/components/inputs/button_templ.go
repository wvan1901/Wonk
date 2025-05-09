// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.793
package inputs

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import "maps"

type ButtonOptions struct {
	Text     string
	Varient  string
	Padding  string
	Disabled bool
	Htmx     HtmxOptions
	OnClick  *string
}

type HtmxOptions struct {
	HxGet     *string
	HxPut     *string
	HxDelete  *string
	HxTarget  *string
	HxSwap    *string
	HxPushUrl *string
	HxTrigger *string
	HxInclude *string
}

func (b *ButtonOptions) TemplAttributes() templ.Attributes {
	tmplAttr := templ.Attributes{}
	if b.Disabled {
		tmplAttr["disabled"] = b.Disabled
	}

	btnClasses := " uppercase font-bold rounded disabled:bg-stone-400 disabled:text-white focus:outline-none"
	paddingClasses := ""
	switch b.Padding {
	case "r1":
		paddingClasses = " py-1 px-2"
	case "r2":
		paddingClasses = " py-2 px-4"
	case "s1":
		paddingClasses = " py-1 px-1"
	case "s2":
		paddingClasses = " py-2 px-2"
	default:
		paddingClasses = " py-2 px-4"
	}
	switch b.Varient {
	case "contained":
		hoverTailwind := " hover:bg-varient-primary-hover"
		focusTailwind := " focus:ring-txt-primary focus:border-txt-primary"
		tmplAttr["class"] = "bg-varient-primary text-bg-main border-2 border-transparent" + hoverTailwind + focusTailwind + btnClasses + paddingClasses
	case "text":
		hoverTailwind := " hover:bg-varient-primary/10"
		focusTailwind := " focus:ring-varient-primary focus:border-varient-primary"
		tmplAttr["class"] = "bg-transparent text-varient-primary border-2 border-transparent" + hoverTailwind + focusTailwind + btnClasses + paddingClasses
	case "outline":
		hoverTailwind := " hover:bg-varient-primary/10"
		focusTailwind := " focus:ring-txt-primary focus:border-txt-primary"
		tmplAttr["class"] = "bg-transparent text-varient-primary border-2 border-varient-primary" + hoverTailwind + focusTailwind + btnClasses + paddingClasses
	default:
		hoverTailwind := " hover:bg-varient-primary-hover"
		focusTailwind := " focus:ring-txt-primary focus:border-txt-primary"
		tmplAttr["class"] = "bg-varient-primary text-bg-main border-2 border-transparent" + hoverTailwind + focusTailwind + btnClasses + paddingClasses
	}

	if b.OnClick != nil {
		tmplAttr["onClick"] = b.OnClick
	}

	htmxAttr := b.Htmx.TemplAttributes()

	maps.Copy(tmplAttr, htmxAttr)

	return tmplAttr
}

func (h *HtmxOptions) TemplAttributes() templ.Attributes {
	tmplAttr := templ.Attributes{}
	if h.HxGet != nil {
		tmplAttr["hx-get"] = h.HxGet
	}
	if h.HxPut != nil {
		tmplAttr["hx-put"] = h.HxPut
	}
	if h.HxDelete != nil {
		tmplAttr["hx-delete"] = h.HxDelete
	}
	if h.HxTarget != nil {
		tmplAttr["hx-target"] = h.HxTarget
	}
	if h.HxSwap != nil {
		tmplAttr["hx-swap"] = h.HxSwap
	}
	if h.HxPushUrl != nil {
		tmplAttr["hx-push-url"] = h.HxPushUrl
	}
	if h.HxTrigger != nil {
		tmplAttr["hx-trigger"] = h.HxTrigger
	}
	if h.HxInclude != nil {
		tmplAttr["hx-include"] = h.HxInclude
	}
	return tmplAttr
}

func Button(opts ButtonOptions) templ.Component {
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
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<button")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ.RenderAttributes(ctx, templ_7745c5c3_Buffer, opts.TemplAttributes())
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</button>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

func ButtonText(opts ButtonOptions) templ.Component {
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
		templ_7745c5c3_Var2 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var2 == nil {
			templ_7745c5c3_Var2 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var3 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
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
			var templ_7745c5c3_Var4 string
			templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(opts.Text)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `app/templates/components/inputs/button.templ`, Line: 112, Col: 13}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return templ_7745c5c3_Err
		})
		templ_7745c5c3_Err = Button(opts).Render(templ.WithChildren(ctx, templ_7745c5c3_Var3), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
