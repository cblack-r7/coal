include config.mk

JNK = coalkit.h 

BIN = coalkit
CLI = coal
KEYGEN = coal-keygen

all: keygen genkeys client backdoor

keygen:
	go build -a -o $(KEYGEN) ./cmd/$(KEYGEN).go

client:
	go build -a -o $(CLI) ./cmd/$(CLI).go

backdoor:
	CC=$(CC) LDFLAGS=$(LDFLAGS) CGO_ENABLED=1 GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags="-s -w $(subst $\",,$(GO_EXT_VARS)) -extldflags $(EXTLDFLAGS)" -a -buildmode=c-shared cmd/$(BIN).go

genkeys:
ifeq (,$(wildcard $(CLI_KF)))
	@printf "generating client keys\n"
	./$(KEYGEN) -o $(CLI_KF)
endif
ifeq (,$(wildcard $(SRV_KF)))
	@printf "generating server keys\n"
	./$(KEYGEN) -o $(SRV_KF)
endif

clean-keys:
	rm -f $(CLI_KF).pub $(CLI_KF) $(SRV_KF).pub $(SRV_KF)

clean:
	rm -f $(JNK) $(BIN).so $(KEYGEN) $(CLI)

.PHONY:
	all clean keygen client backdoor genkeys clean-keys
