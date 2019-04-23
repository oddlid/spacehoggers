# Spacehoggers

A rewrite of `spacehoggers.pl` in pure Go, as the Perl script has been quite useful, but installing extra Perl modules is annoying.
A static binary is good for this utility.

The field "Size" reports the actual byte size of the file(s). The field "Usage" reports the blocks * block size (512) used, giving the disk usage. The field "Path" is, well, the relative path.

## Usage:

    $ spacehoggers -h
    NAME:
       spacehoggers - Find biggest/smallest files/dirs

    USAGE:
       spacehoggers [global options] command [command options] [arguments...]

    VERSION:
       2019-04-23_968e079 (Compiled: 2019-04-23T20:52:38+02:00)

    AUTHOR:
       Odd E. Ebbesen <oddebb@gmail.com>

    COMMANDS:
         help, h  Shows a list of commands or help for one command

    GLOBAL OPTIONS:
       --root DIR, -R DIR        DIR to check (default: ".")
       --all, -a                 List all files instead of summarizing directories
       --sort OPTION, -s OPTION  Sort by OPTION: size or usage (default: "size")
       --reverse, -r             Reverse order (smallest to largest)
       --limit value, -l value   How many results to display (default: 10)
       --log-level level         Log level (options: debug, info, warn, error, fatal, panic) (default: "info")
       --debug, -d               Run in debug mode [$DEBUG]
       --help, -h                show help
       --version, -v             print the version

    COPYRIGHT:
       (c) 2019 Odd Eivind Ebbesen
