#!/bin/bash

function install {
    echo "Installing pre-commit script"
    
    if [[ -f .git/hooks/pre-commit ]]; then 
        read -p 'Pre-commit hook already exists. Remove? (Y/N): ' confirm
        if [[ "$confirm" != "Y" ]]; then
            echo "Aborting installation." 
            exit 1
        else
            echo "Removing existing pre-commit script"
            rm .git/hooks/pre-commit || (echo "Failed to remove script"; exit 1) && echo "Removed existing script"
        fi
    fi
    
    echo "Installing pre-commit hook"
    ln pre-commit.sh .git/hooks/pre-commit
}

function pre-commit {
    echo "Running pre-commit script"

    if [[ -f venv/bin/activate ]]; then
        . venv/bin/activate
    else 
        echo "Running without a virtual environment" >&2
    fi

    make lint
    lint_return_code=$?
    make test
    test_return_code=$?

    echo "---- Pre-commit result:"

    if [ $test_return_code != 0 ]; then
        echo "Tests failed"
    else
        echo "Tests succeeded"
    fi

    if [ $lint_return_code != 0 ]; then
        echo "Linters failed"
    else
        echo "Linters succeeded"
    fi

    if [ $test_return_code != 0 ] || [ $lint_return_code != 0 ]; then
        exit 1
    fi
}

if [ "$1" = "install" ]; then
    install
else
    pre-commit
fi