.PHONY: build test test-unit test-verbose coverage coverage-html mutation mutation-docker mutation-app clean

BINARY := njord
PKG := ./...
MUTATION_PKGS := ./internal/app/ ./internal/docker/ ./internal/config/

build:
	go build -o $(BINARY) ./cmd/njord/

# test: roda unit tests + mutation testing (threshold 84%)
# Se a eficácia cair abaixo de 84%, bloqueia (exit != 0).
# Pacotes sem mutantes Killed+Lived são ignorados (não há o que medir).
MUTATION_THRESHOLD := 84

test: test-unit
	@echo "→ Mutation testing (threshold eficácia=$(MUTATION_THRESHOLD)%)"
	@fail=0; \
	for pkg in $(MUTATION_PKGS); do \
		echo "  $$pkg"; \
		out=$$(gremlins unleash $$pkg 2>&1); \
		killed=$$(echo "$$out" | grep -oE 'Killed: [0-9]+' | grep -oE '[0-9]+'); \
		lived=$$(echo "$$out" | grep -oE 'Lived: [0-9]+' | grep -oE '[0-9]+'); \
		timedout=$$(echo "$$out" | grep -oE 'Timed out: [0-9]+' | grep -oE '[0-9]+'); \
		effective_killed=$$((killed + timedout)); \
		total=$$((effective_killed + lived)); \
		if [ "$$total" -eq 0 ]; then \
			echo "    ↳ sem mutantes testáveis — pulado"; \
			continue; \
		fi; \
		efficacy=$$(LC_ALL=C awk -v k=$$effective_killed -v t=$$total 'BEGIN{printf "%.2f", (k/t)*100}'); \
		echo "    ↳ eficácia: $$efficacy% (killed=$$killed timedout=$$timedout lived=$$lived)"; \
		below=$$(LC_ALL=C awk -v e=$$efficacy -v t=$(MUTATION_THRESHOLD) 'BEGIN{print (e+0<t+0)?1:0}'); \
		if [ "$$below" -eq 1 ]; then \
			echo "    ✗ ABAIXO do threshold $(MUTATION_THRESHOLD)% — fortalecer testes"; \
			fail=1; \
		fi; \
	done; \
	if [ $$fail -ne 0 ]; then \
		echo "❌ Mutation testing reprovou. Corrija os testes antes de commitar."; \
		exit 1; \
	fi; \
	echo "✓ Todos os pacotes passaram no threshold de $(MUTATION_THRESHOLD)%"

# test-unit: só unit tests, para dev loop rápido
test-unit:
	go test $(PKG)

test-verbose:
	go test -v $(PKG)

coverage:
	go test -cover $(PKG)

coverage-html:
	go test -coverprofile=coverage.out $(PKG)
	go tool cover -html=coverage.out -o coverage.html
	@echo "Relatório: coverage.html"

mutation:
	gremlins unleash $(PKG)

mutation-docker:
	gremlins unleash ./internal/docker/

mutation-app:
	gremlins unleash ./internal/app/

clean:
	rm -f $(BINARY) coverage.out coverage.html
