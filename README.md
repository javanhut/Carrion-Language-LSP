Below is a complete **README.md** that you can place in your repository. It assumes your project is called **Carrion Language LSP** and explains how to build, install, and configure it in Neovim with **nvim-chad**.

```markdown
# Carrion Language Server

A custom [Language Server Protocol (LSP)](https://microsoft.github.io/language-server-protocol/) implementation for the **Carrion** programming language, built in Go.

## Features

- Basic LSP initialization (connect, shutdown, etc.)  
- Easy to extend with custom handlers (e.g. diagnostics, hover, completions)

> **Note**: This is a minimal proof-of-concept right now. You can add more LSP features in `main.go` or related handlers (e.g., diagnostics, hover, definition, etc.).

---

## Prerequisites

1. [Go 1.18+](https://go.dev/dl/) installed and in your `PATH`.
2. A working Neovim (v0.7 or later) setup with the [nvim-lspconfig](https://github.com/neovim/nvim-lspconfig) plugin.  
3. (Recommended) The [nvim-chad](https://github.com/NvChad/NvChad) configuration framework if you want the same layout discussed here.

---

## Installation

1. **Clone this repository**:
   ```bash
   git clone https://github.com/javanhut/Carrion-Language-LSP.git
   cd Carrion-Language-LSP
   ```

2. **Build the Go binary**:
   ```bash
   go mod tidy
   go build -o carrion-lsp main.go
   ```

3. **Make it executable (on Linux/macOS)**:
   ```bash
   chmod +x carrion-lsp
   ```

4. **(Optional) Move it into your PATH**  
   ```bash
   mv carrion-lsp ~/go/bin/   # or /usr/local/bin/ or another folder in $PATH
   ```
   If you don’t move it, be sure to reference the absolute path in your Neovim config.

---

## Neovim Configuration (using nvim-chad)

Below is an **example** setup for nvim-chad (v2). Adapt file names/paths as needed.

### 1. Create a Filetype Definition

In `~/.config/nvim/lua/custom/filetypes.lua`, associate your Carrion files (e.g., `.crl`) with the `carrion` filetype:

```lua
vim.api.nvim_create_autocmd({ "BufRead", "BufNewFile" }, {
  pattern = "*.crl",  -- or whatever extension(s) you use
  callback = function()
    vim.bo.filetype = "carrion"
  end,
})
```

### 2. Create a Carrion LSP Server Definition

Create a file at:
```
~/.config/nvim/lua/custom/lspconfig/servers/carrion_language_lsp.lua
```
with the following content:

```lua
local M = {}

M.config = function()
  local on_attach = require("plugins.configs.lspconfig").on_attach
  local capabilities = require("plugins.configs.lspconfig").capabilities

  local lspconfig = require("lspconfig")
  local configs = require("lspconfig.configs")
  local util = require("lspconfig/util")

  -- 1) Define "carrion_language_lsp" if not already defined (for older nvim-lspconfig)
  if not configs.carrion_language_lsp then
    configs.carrion_language_lsp = {
      default_config = {
        cmd = { "/absolute/path/to/carrion-lsp" },  -- or "carrion-lsp" if it's in $PATH
        filetypes = { "carrion" },
        root_dir = function(fname)
          return util.root_pattern(".git", "carrion.config")(fname)
            or util.path.dirname(fname)
        end,
        single_file_support = true,
      },
    }
  end

  -- 2) Now set up the server
  lspconfig.carrion_language_lsp.setup {
    on_attach = on_attach,
    capabilities = capabilities,
    cmd = { "/absolute/path/to/carrion-lsp" }, -- Must match above or override
    filetypes = { "carrion" },
    root_dir = util.root_pattern(".git", "carrion.config"),
    settings = {
      -- If your LSP supports custom settings, define them here
    },
  }
end

return M
```

> **Note**: If your binary is not in the system `PATH`, replace `"/absolute/path/to/carrion-lsp"` with, for example, `"/home/username/Carrion-Language-LSP/carrion-lsp"`.

### 3. Load the Carrion LSP Config

In nvim-chad, you typically have a file named `~/.config/nvim/lua/custom/configs/lspconfig.lua`. Add a line requiring your Carrion LSP:

```lua
-- ~/.config/nvim/lua/custom/configs/lspconfig.lua
local on_attach = require("plugins.configs.lspconfig").on_attach
local capabilities = require("plugins.configs.lspconfig").capabilities
local lspconfig = require("lspconfig")

-- Example: existing gopls setup
lspconfig.gopls.setup {
  on_attach = on_attach,
  capabilities = capabilities,
  -- ...
}

-- Now require Carrion's config
require("custom.lspconfig.servers.carrion_language_lsp").config()
```

Also, in your `~/.config/nvim/lua/custom/chadrc.lua` (or wherever nvim-chad loads servers), ensure:
```lua
---@type ChadrcConfig
local M = {}
M.ui = { theme = "catppuccin" }
M.plugins = "custom.plugins"
M.mappings = require "custom.mappings"

M.lsp = {
  servers = {
    "carrion_language_lsp",
    -- ...
  },
}

return M
```
### 4. Make a filedetect map for This

Create Filetype Detection for Carrion Files

    Create the ftdetect Directory (if it doesn’t exist):
    ```bash
    mkdir -p ~/.config/nvim/ftdetect
```

Create the File carrion.vim:

```bash
nvim ~/.config/nvim/ftdetect/carrion.vim
```

Add the Following Line and Save:

" Set filetype to carrion for files ending in .crl
au BufRead,BufNewFile *.crl set filetype=carrion


### 5. Verify It Works

1. Open a file that ends with `.crl`.  
2. Run `:set filetype?` — you should see `carrion`.  
3. Run `:LspInfo` — you should see:
   ```
   LSP configs active in this session (globally):
   - Configured servers: carrion_language_lsp, ...

   LSP configs for this buffer:
   - 1 client(s) attached: carrion_language_lsp
   ...
   ```
4. If it’s not attaching, check:
   - Path to the `carrion-lsp` binary.
   - Whether you have `.git` or `carrion.config` in your project folder (if `root_dir` relies on those).
   - The Neovim LSP logs (`:echo stdpath("cache")` → `lsp.log`).

---

## Usage & Customization

Once it’s running, you can add custom LSP functionality in Go by implementing more handlers (diagnostics, hover, completions, etc.) in the `main.go` file (or additional Go files). Neovim will automatically communicate with these new endpoints via LSP requests/responses.

---

## Contributing

1. **Fork** the repository.  
2. **Create** a feature branch (`git checkout -b feature/my-feature`).  
3. **Commit** your changes.  
4. **Push** the branch (`git push origin feature/my-feature`).  
5. **Open** a Pull Request on GitHub.

---

**Thanks for using Carrion Language LSP!** If you encounter any issues or have feature requests, please open an issue on this repository. Happy Coding!
```
