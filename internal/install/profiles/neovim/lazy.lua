local lazypath = vim.fn.stdpath("data") .. "/lazy"
vim.opt.rtp:prepend(lazypath .. "/lazy.nvim")

require("lazy").setup({
  { dir = lazypath .. "/LazyVim", name = "LazyVim", import = "lazyvim.plugins", lazy = false },
  { dir = lazypath .. "/nvim-treesitter", name = "nvim-treesitter", lazy = false },
  { import = "plugins" },
}, {
  defaults = { lazy = true, version = false },
  checker = { enabled = false },
  change_detection = { notify = false },
})
