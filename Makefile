#Go's build system is nice, but makefiles are nicer
#
#

BUILDIR := obj
GOBIN := go
GOCMD := $(GOBIN)

GO_SRC = $(wildcard src/*.go)
GO_OBJ := $(BUILDIR)/extra

.PHONY : all $(GO_OBJ).elf

all: $(GO_OBJ).elf

$(GO_OBJ).elf: $(GO_SRC)
	@echo "+GO "$(GO_SRC)
	@$(GOCMD) build -o $@ $(GO_SRC)
