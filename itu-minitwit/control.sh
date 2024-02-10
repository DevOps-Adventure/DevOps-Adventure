#!/usr/bin/env bash
if [ "$1" = "init" ]; then

    if [ -f "./tmp/minitwit.db" ]; then 
        echo "Database already exists."
        exit 1
    fi
    echo "Putting a database to /tmp/minitwit.db..."
    python3 -c"from minitwit import init_db;init_db()"
elif [ "$1" = "start" ]; then
    echo "Starting minitwit..."
    nohup "$(which python3)" minitwit.py > /tmp/out.log 2>&1 &
elif [ "$1" = "stop" ]; then
    echo "Stopping minitwit..."
    pkill -f minitwit
elif [ "$1" = "inspectdb" ]; then
    if [ "$(uname -s)" == "Darwin" ]; then #MacOS
        if [ "$(uname -m)" == "x86_64" ]; then #intel
            ./MACIntel_flag_tool.out -i | less
        elif [ "$(uname -m)" == "arm64" ]; then #M1
            ./MAC_flag_tool.out -i | less 
        else
            echo "Unsupported architecture: $(uname -m)"
            exit 1
        fi
    else
        ./flag_tool -i | less
    fi
elif [ "$1" = "flag" ]; then
    if [ "$(uname -s)" == "Darwin" ]; then #MacOS
        if [ "$(uname -m)" == "x86_64" ]; then #intel
            ./MACIntel_flag_tool.out "$@"
        elif [ "$(uname -m)" == "arm64" ]; then #M1
            ./MAC_flag_tool.out "$@"
        else
            echo "Unsupported architecture: $(uname -m)"
            exit 1
        fi
    else
        ./flag_tool "$@"
fi
else
  echo "I do not know this command..."
fi