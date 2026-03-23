# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

load("@protobuf//bazel/common:proto_info.bzl", "ProtoInfo")

"""Bazel rule for generating TypeScript bindings from proto_library targets
using the ts-proto (protoc-gen-ts_proto) plugin.

Usage in a BUILD.bazel file:

    load("//proto:ts_proto_rules.bzl", "ts_proto_gen")

    ts_proto_gen(
        name = "ts_proto_files",
        proto = ":my_proto",
        visibility = ["//visibility:public"],
    )
"""

def _ts_proto_gen_impl(ctx):
    p_info = ctx.attr.proto[ProtoInfo]
    direct_srcs = p_info.direct_sources

    if not direct_srcs:
        return [DefaultInfo(files = depset([]))]

    outputs = []
    for src in direct_srcs:
        # Declare output .ts file in the same package directory.
        # ctx.actions.declare_file(name) creates a file at:
        #   bazel-out/<config>/bin/<package>/<name>
        # protoc with --proto_path=. writes:
        #   {ts_proto_out}/<proto_package_path>/<basename>.ts
        # e.g. for proto/api/v1/foo.proto → bazel-out/.../bin/proto/api/v1/foo.ts
        # Both paths agree when ts_proto_out = ctx.bin_dir.path.
        ts_file = ctx.actions.declare_file(
            src.basename.removesuffix(".proto") + ".ts",
        )
        outputs.append(ts_file)

    if not outputs:
        return [DefaultInfo(files = depset([]))]

    # Build --proto_path flags from proto_info so that all transitive imports
    # (googleapis annotations, well-known types, etc.) resolve correctly in the
    # Bazel sandbox without any hard-coded external-repo paths.
    proto_path_args = []
    for path in p_info.transitive_proto_path.to_list():
        proto_path_args.append("--proto_path=" + path)

    plugin = ctx.executable._plugin

    args = ctx.actions.args()
    args.add_all(proto_path_args)
    args.add("--plugin=protoc-gen-ts_proto=" + plugin.path)

    # Output directory must match where declare_file places files.
    # For a rule in package "proto/api/v1", ctx.bin_dir.path is
    # "bazel-out/<config>/bin" and declare_file("foo.ts") lands at
    # "bazel-out/<config>/bin/proto/api/v1/foo.ts" which matches
    # protoc's output of "{ts_proto_out}/proto/api/v1/foo.ts".
    args.add("--ts_proto_out=" + ctx.bin_dir.path)
    args.add(
        "--ts_proto_opt=" +
        "esModuleInterop=true," +
        "forceLong=long," +
        "useOptionals=messages," +
        "outputClientImpl=grpc-web",
    )
    for src in direct_srcs:
        args.add(src.path)

    ctx.actions.run(
        mnemonic = "TsProtoGen",
        progress_message = "Generating TypeScript proto bindings for %s" % ctx.label,
        inputs = depset(transitive = [p_info.transitive_sources]),
        outputs = outputs,
        executable = ctx.executable._protoc,
        arguments = [args],
        tools = [plugin],
        env = {"BAZEL_BINDIR": ctx.bin_dir.path},
    )

    return [DefaultInfo(files = depset(outputs))]

ts_proto_gen = rule(
    implementation = _ts_proto_gen_impl,
    attrs = {
        "proto": attr.label(
            mandatory = True,
            providers = [ProtoInfo],
            doc = "The proto_library target to generate TypeScript for.",
        ),
        "_protoc": attr.label(
            default = Label("@protobuf//:protoc"),
            executable = True,
            cfg = "exec",
            doc = "The protoc compiler binary.",
        ),
        "_plugin": attr.label(
            default = Label("//ui:ts_proto_plugin"),
            executable = True,
            cfg = "exec",
            doc = "The protoc-gen-ts_proto plugin binary (from the ts-proto npm package).",
        ),
    },
    doc = "Generates TypeScript bindings from a proto_library target using ts-proto.",
)

def _ts_wkt_gen_impl(ctx):
    """Generates TypeScript bindings for WKT (well-known type) proto_library targets.

    Unlike ts_proto_gen (which uses src.basename for the output path), this rule
    preserves the proto package path in the output (e.g. google/protobuf/timestamp.ts).
    It derives the output sub-path by stripping the proto_path prefix from each
    source file's full path.
    """
    plugin = ctx.executable._plugin
    protoc = ctx.executable._protoc
    all_outputs = []

    for proto_target in ctx.attr.protos:
        p_info = proto_target[ProtoInfo]

        proto_path_args = []
        for path in p_info.transitive_proto_path.to_list():
            proto_path_args.append("--proto_path=" + path)

        for src in p_info.direct_sources:
            # Find which proto_path prefix this source belongs to so we can
            # compute the output file path relative to that root.
            rel_path = None
            for path in p_info.transitive_proto_path.to_list():
                if src.path.startswith(path + "/"):
                    rel_path = src.path[len(path) + 1:]
                    break
            if rel_path == None:
                rel_path = src.basename

            ts_rel_path = rel_path.removesuffix(".proto") + ".ts"
            ts_file = ctx.actions.declare_file(ts_rel_path)
            all_outputs.append(ts_file)

            # Use the bin root so protoc writes the file directly to
            # bazel-out/{config}/bin/{rel_path}.ts, which matches the path
            # that ts-proto generates relative imports against.
            out_dir = ctx.bin_dir.path

            args = ctx.actions.args()
            args.add_all(proto_path_args)
            args.add("--plugin=protoc-gen-ts_proto=" + plugin.path)
            args.add("--ts_proto_out=" + out_dir)
            args.add(
                "--ts_proto_opt=" +
                "esModuleInterop=true," +
                "forceLong=long," +
                "useOptionals=messages," +
                "outputClientImpl=grpc-web",
            )
            args.add(src.path)

            ctx.actions.run(
                mnemonic = "TsWktGen",
                progress_message = "Generating TypeScript WKT bindings for %s" % src.basename,
                inputs = depset(transitive = [p_info.transitive_sources]),
                outputs = [ts_file],
                executable = protoc,
                arguments = [args],
                tools = [plugin],
                env = {"BAZEL_BINDIR": ctx.bin_dir.path},
            )

    return [DefaultInfo(files = depset(all_outputs))]

ts_wkt_gen = rule(
    implementation = _ts_wkt_gen_impl,
    attrs = {
        "protos": attr.label_list(
            mandatory = True,
            providers = [ProtoInfo],
            doc = "List of proto_library targets (WKTs) to generate TypeScript for.",
        ),
        "_protoc": attr.label(
            default = Label("@protobuf//:protoc"),
            executable = True,
            cfg = "exec",
        ),
        "_plugin": attr.label(
            default = Label("//ui:ts_proto_plugin"),
            executable = True,
            cfg = "exec",
        ),
    },
    doc = "Generates TypeScript bindings from WKT proto_library targets, preserving package paths.",
)
