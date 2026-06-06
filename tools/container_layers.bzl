def _go_api_layer_impl(ctx):
    out = ctx.actions.declare_file(ctx.label.name + ".tar")
    args = ctx.actions.args()
    args.add(out)
    args.add(ctx.file.go)
    args.add_all(ctx.files.srcs)

    ctx.actions.run_shell(
        inputs = ctx.files.srcs,
        tools = depset(ctx.files.sdk + [ctx.file.go]),
        outputs = [out],
        arguments = [args],
        command = """
set -euo pipefail
out="$1"
go_tool="$2"
shift 2
execroot="$PWD"
work="$(mktemp -d "${TMPDIR:-/tmp}/go-api-layer.XXXXXX")"
trap 'chmod -R u+w "$work" 2>/dev/null || true; rm -rf "$work"' EXIT

for src in "$@"; do
  mkdir -p "$work/$(dirname "$src")"
  cp -L "$src" "$work/$src"
done

cd "$work"
go_tool="$execroot/$go_tool"
export GOROOT="$(dirname "$(dirname "$go_tool")")"
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export GOCACHE="$work/.gocache"
export GOENV=off
export GOPATH="$work/.gopath"
export GOTOOLCHAIN=local

"$go_tool" build -trimpath -ldflags="-s -w" -o hello-app-fork-api ./api
tar --mtime="UTC 1970-01-01" --owner=0 --group=0 --numeric-owner -cf "$execroot/$out" hello-app-fork-api
""",
        mnemonic = "GoApiLayer",
        progress_message = "Building Go API layer %{label}",
    )

    return DefaultInfo(files = depset([out]))

go_api_layer = rule(
    implementation = _go_api_layer_impl,
    attrs = {
        "go": attr.label(
            allow_single_file = True,
            cfg = "exec",
            mandatory = True,
        ),
        "sdk": attr.label(
            allow_files = True,
            mandatory = True,
        ),
        "srcs": attr.label_list(
            allow_files = True,
            mandatory = True,
        ),
    },
)

def _web_static_layer_impl(ctx):
    out = ctx.actions.declare_file(ctx.label.name + ".tar")
    args = ctx.actions.args()
    args.add(out)
    args.add(ctx.file.dist.path)
    args.add(ctx.file.nginx_conf)

    ctx.actions.run_shell(
        inputs = [ctx.file.dist, ctx.file.nginx_conf],
        outputs = [out],
        arguments = [args],
        command = """
set -euo pipefail
out="$1"
dist="$2"
nginx_conf="$3"
execroot="$PWD"
root="$(mktemp -d "${TMPDIR:-/tmp}/web-layer.XXXXXX")"
trap 'rm -rf "$root"' EXIT

mkdir -p "$root/usr/share/nginx/html" "$root/etc/nginx/conf.d"
cp -R "$dist"/. "$root/usr/share/nginx/html/"
cp "$nginx_conf" "$root/etc/nginx/conf.d/default.conf"
cd "$root"
tar --mtime="UTC 1970-01-01" --owner=0 --group=0 --numeric-owner -cf "$execroot/$out" .
""",
        mnemonic = "WebStaticLayer",
        progress_message = "Building web static layer %{label}",
    )

    return DefaultInfo(files = depset([out]))

web_static_layer = rule(
    implementation = _web_static_layer_impl,
    attrs = {
        "dist": attr.label(
            allow_single_file = True,
            mandatory = True,
        ),
        "nginx_conf": attr.label(
            allow_single_file = True,
            mandatory = True,
        ),
    },
)
