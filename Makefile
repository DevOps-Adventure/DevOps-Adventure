DIR := go-minitwit
TIMEOUT := 10s
SERVER_STARTS := 3s

STORE_TEST_RESULT_CMD_SUFFIX := && echo "success" > .test.txt || && echo "failure" > .test.txt
STORE_LINT_RESULT_CMD_SUFFIX := && echo "success" > .lint.txt || && echo "failure" > .lint.txt

# Colors for echo
CYAN := \033[0;36m
YELLOW := \033[1;33m
RESET := \033[0m

help:
	@echo "------- H e L P  u.u ------¬"
	@echo "    $(YELLOW)dep                       $(CYAN)Install all dependencies"
	@echo "    $(YELLOW)format                    $(CYAN)Format sources"
	@echo "    $(YELLOW)run                       $(CYAN)Run the application"
	@echo "    $(YELLOW)lint                      $(CYAN)Run linters"
	@echo "    $(YELLOW)test                      $(CYAN)Run tests"
	@echo "    $(YELLOW)help                      $(CYAN)Show help$(RESET)"
	@echo "    $(YELLOW)pre-commit_install        $(CYAN)Install pre-commit hook$(RESET)"
	@echo "    $(YELLOW)pre-commit_uninstall      $(CYAN)Uninstall pre-commit hook$(RESET)"

run:
	@echo "$(CYAN)Running the service$(RESET)"
	@cd $(DIR) && go run *.go > /dev/null

run_bg:
	@echo "$(CYAN)Service timeout set:$(YELLOW) $(TIMEOUT)$(RESET)"
	@(cd $(DIR) > /dev/null && ( \
		go run *.go > /dev/null 2>&1 & echo $$! > .pidfile ; \
		# Tracks the process PID -> echo "PID of the process: $$(cat .pidfile)"; \
		sleep $(TIMEOUT); \
		# kills the process -> echo "Killing process with PID $$(cat .pidfile)"; \
		kill -9 $$(cat .pidfile) > /dev/null; \
		# echo "Process killed"; \
		rm .pidfile \
	)) &


format:
	@echo "$(CYAN)Formatting source tree$(RESET)"
	@cd $(DIR) && go fmt ./...

lint:
	@echo "$(CYAN)Running CLI linters$(RESET)"
	@cd $(DIR) && touch .lint.txt
	@cd $(DIR) && golangci-lint run && echo "success" > .lint.txt || echo "failure" > .lint.txt
	@cd $(DIR) && gofumpt -l -w . && echo "success" > .lint.txt || echo "failure" > .lint.txt
	@cd $(DIR) grep "failure" .lint.txt && rm .lint.txt && exit 1 || rm .lint.txt exit 0

test:
	@if [ -z "${VIRTUAL_ENV}" ]; then \
		echo "$(CYAN)--- $(YELLOW)Warning:$(CYAN) Not testing for $(YELLOW)python-sdk$(CYAN) without an activated virtual environment$(RESET)"; exit 1; \
	else \
		echo "$(CYAN)Starting the server in the $(YELLOW)background$(RESET)";\
		$(MAKE) run_bg > /dev/null;\
		echo "$(CYAN)Waiting for server to be ready...$(RESET)";\
		sleep $(SERVER_STARTS); \
		echo "$(CYAN)Running tests against $(YELLOW)go-minitwit$(RESET)"; \
		cd $(DIR) && touch .test.txt; \
		pytest tests && echo "success" > .test.txt || echo "failure" > .test.txt; \
		grep "failure" .test.txt && rm .test.txt && exit 1 || rm .test.txt && exit 0; \
	fi

dep:
	@echo "$(CYAN)Checking the dependencies of $(YELLOW)go-minitwit$(RESET)"
	@cd $(DIR) && go mod tidy  > /dev/null

pre-commit_install:
	./pre-commit.sh install

pre-commit_uninstall:
	@rm -f .git/hooks/pre-commit
	@echo "$(CYAN)Pre-commit script $(YELLOW)removed!$(RESET)"
