# Claude Code Context Switcher

## Background
As a claude-code user, I often need to switch the model provider to use this program by setting two key environment variables: `ANTHROPIC_BASE_URL` and `ANTHROPIC_AUTH_TOKEN` .  
Currently how do I do the switching is to set these two variables inside a environment file. After modifying the file, I use `source <environment_file>` to make these two environment vairables take effect. 
However, this switch is very clumsy and it involves a lot of manual operations.  As a result, I want to develop a CLI application that can help me quickly switch these two environment variables in the current shell session.  In this way, after I run claude-code next time, it will direct take effect

## Application Requirement
- The program name is ccctx
- A context means a pair of environment variable values for `ANTHROPIC_BASE_URL` and `ANTHROPIC_AUTH_TOKEN`
- The user create a configuration file ( TOML format ) to set a list of contexts in a specific folder ( maybe $HOME/.ccctx/config.toml ) .  The file format might be in this way:
  ```toml
  [context.ctx1]
  base_url = XXXXX
  auth_token = YYYYY

  [context.ctx2]
  base_url = AAAAA
  auth_token = BBBBB
  ```
- The user cause use `ccctx` with sub-commands:
    - `ccctx` or `ccctx --help` or `ccctx -h`: show help
    - `ccctx list` :  show the currently available contexts in the config file
    - `ccctx switch [ctx-name]` :  
      - If not specify a ctx-name , it will open up a interactive cursor than can move up / down , and after the user hit ENTER, the highlighted context will be chosen as the context
      - If specify a ctx-name, then direct set that name as the context
      - Setting the context means after the program is exited, the shell running `ccctx` has changed the environment variable `ANTHROPIC_BASE_URL` and `ANTHROPIC_AUTH_TOKEN`, so that in the next time when calling claude-code, the new environment varialbes values will take effect
    - `ccctx run [ctx-name]` : directly run the claude code with the environment variables with the specified context set.  After running, the environment varialbes will be the ones before calling the `ccctx`.  If ccctx-name is not specified, it will open up a interactive cursor that user can move up / down and choose the context he wants to use, and hit ENTER
