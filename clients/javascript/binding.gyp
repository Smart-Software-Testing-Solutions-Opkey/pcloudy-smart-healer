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
          ["OS=='linux'", {
              "libraries": ["<(module_root_dir)/lib/linux-x64/libsmarthealer.a"]
          }]
      ]
    }
  ]
}
