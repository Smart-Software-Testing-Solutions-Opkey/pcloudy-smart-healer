{
  "targets": [
    {
      "target_name": "smarthealer",
      "sources": ["src/smarthealer.cc"],
      "include_dirs": [
        "<!(node -p \"require('node-addon-api').include\")",
        "includes"
      ],
      "dependencies": [
        "<!(node -p \"require('node-addon-api').targets\"):node_addon_api"
      ],
      "cflags!": [ "-fno-exceptions" ],
      "cflags_cc!": [ "-fno-exceptions" ],
      "cflags": [ "-fexceptions" ],
      "cflags_cc": [ "-fexceptions" ],
      "conditions": [
        ["OS=='linux' and target_arch=='x64'", {
          "libraries": ["<(module_root_dir)/lib/linux-amd64/libsmarthealer.a"]
        }],
        ["OS=='linux' and target_arch=='arm64'", {
          "libraries": ["<(module_root_dir)/lib/linux-arm64/libsmarthealer.a"]
        }],
        ["OS=='mac' and target_arch=='x64'", {
          "libraries": ["<(module_root_dir)/lib/darwin-amd64/libsmarthealer.a"]
        }],
        ["OS=='mac' and target_arch=='arm64'", {
          "libraries": ["<(module_root_dir)/lib/darwin-arm64/libsmarthealer.a"]
        }]
      ]
    }
  ]
}
