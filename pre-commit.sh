#!/bin/bash

RED='\033[1;31m'
GREEN='\033[1;32m'
MAGENTA='\033[1;35m'
ORANGE='\033[0;33m'
RESET='\033[0m'

function install {
    echo -e "${MAGENTA}Installing pre-commit script ${RESET}"
    
    if [[ -f .git/hooks/pre-commit ]]; then 
        read -p $'\033[0;33m Pre-commit hook already exists. Remove? (Y/N): \033[0m' confirm
        if [[ "$confirm" != "Y" ]]; then
            echo -e "${RED}Aborting installation.${RESET}" 
            exit 1
        else
            echo -e "${MAGENTA}Removing existing pre-commit script${RESET}"
            rm .git/hooks/pre-commit || (echo -e "${RED}Failed to remove script${RESET}"; exit 1) && echo -e "${GREEN}Existing script removed${RESET}"
        fi
    fi
    
    echo -e "${MAGENTA}Installing pre-commit hook${RESET}"
    ln pre-commit.sh .git/hooks/pre-commit
    echo -e "${GREEN}Pre-commit ready!${RESET}"

}

function pre-commit {
    echo -e "${MAGENTA}--------------------~ Running pre-commit script ~--------------------${RESET}"

    if [[ -f venv/bin/activate ]]; then
        . venv/bin/activate
    else 
        echo -e "${ORANGE}You are trying to run without a virtual environment.${RESET}" >&2
        echo -e "${RED}Aborting commit procedure!${RESET}" 
            exit 1
    fi

    make lint
    lint_return_code=$?
    make test
    test_return_code=$?

    echo -e "${ORANGE}~> ---- Pre-commit result ---- <~${RESET}"

    if [ $lint_return_code != 0 ]; then
        echo -e "${RED}-> Linters failed     ( ✗ )${RESET}"
    else
        echo -e "${GREEN}-> Linters succeeded! ( ✓ )${RESET}"
    fi

    if [ $test_return_code != 0 ]; then
        echo -e "${RED}-> Tests failed       ( ✗ )${RESET}"
    else
        echo -e "${GREEN}-> Tests succeeded!   ( ✓ )${RESET}"
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