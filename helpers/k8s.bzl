_TEMPLATE = "//k8s:deploy.yaml"

def _template_manifest_impl(ctx):
  name = '{}'.format(ctx.label).replace("//cmd/", "").replace("/", "-").replace(":manifest", "")
  ctx.actions.expand_template(
    template = ctx.file._template,
    output = ctx.outputs.source_file,
    substitutions = {
      "{NAME}": name,
    },
  )

template_manifest = rule(
  implementation = _template_manifest_impl,
  attrs = {
    "_template": attr.label(
      default = Label(_TEMPLATE),
      allow_single_file = True,
    ),
  },
  outputs = {"source_file": "%{name}.yaml"},
)

def template_image(ctx, *args, **kwargs):
  print(ctx, args, kwargs)