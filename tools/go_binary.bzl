def _go_api_binary_impl(ctx):
    out = ctx.actions.declare_file(ctx.label.name)
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
work="$(mktemp -d "${TMPDIR:-/tmp}/go-api-binary.XXXXXX")"
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

"$go_tool" build -trimpath -ldflags="-s -w" -o "$execroot/$out" ./api
""",
        mnemonic = "GoApiBinary",
        progress_message = "Building Go API binary %{label}",
    )

    return DefaultInfo(files = depset([out]), executable = out)

go_api_binary = rule(
    implementation = _go_api_binary_impl,
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
    executable = True,
)
