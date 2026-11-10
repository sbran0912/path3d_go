BINARY  := path3d
BIN_DIR := bin
APP_DIR := ./app

.PHONY: build run clean

# Kompiliert die Anwendung nach bin/$(BINARY)
build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(APP_DIR)

# Startet die Anwendung direkt über go run
run:
	go run $(APP_DIR)

# Entfernt die erzeugten Artefakte
clean:
	rm -rf $(BIN_DIR)
