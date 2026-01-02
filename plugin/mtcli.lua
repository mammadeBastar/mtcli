-- mtcli.nvim plugin loader
-- This file is automatically loaded by Neovim

-- Prevent loading twice
if vim.g.loaded_mtcli then
  return
end
vim.g.loaded_mtcli = true

-- Create the command immediately (setup() can be called later for config)
vim.api.nvim_create_user_command('MtType', function()
  -- Ensure setup has been called
  local mtcli = require('mtcli')
  if not mtcli.config then
    mtcli.setup({})
  end
  mtcli.start()
end, { desc = 'Start typing test on function under cursor' })

