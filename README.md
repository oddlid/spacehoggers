# Spacehoggers

A rewrite of `spacehoggers.pl` in pure Go, as the Perl script has been quite useful, but installing extra Perl modules is annoying.
A static binary is good for this utility.

This utility will not take into account filesystem block size, like `du` does, but will only report on the actual sizes of files.

## Usage:

  $ spacehoggers -h
  NAME:
     spacehoggers - Find biggest/smallest files/dirs
  
  USAGE:
     spacehoggers [global options] command [command options] [arguments...]
  
  VERSION:
     2019-01-07_6bee33d (Compiled: 2019-01-07T00:58:46+01:00)
  
  AUTHOR:
     Odd E. Ebbesen <oddebb@gmail.com>
  
  COMMANDS:
       help, h  Shows a list of commands or help for one command
  
  GLOBAL OPTIONS:
     --log-level level        Log level (options: debug, info, warn, error, fatal, panic) (default: "info")
     --debug, -d              Run in debug mode [$DEBUG]
     --root DIR               DIR to check (default: ".")
     --limit value, -l value  How many results to display (default: 10)
     --reverse, -r            Reverse order (smallest to largest)
     --all, -a                List all files instead of summarizing directories
     --help, -h               show help
     --version, -v            print the version
  
  COPYRIGHT:
     (c) 2019 Odd Eivind Ebbesen
