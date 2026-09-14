.PHONY: run test migrate

run:
	$(MAKE) -C backend run

test:
	$(MAKE) -C backend test

migrate:
	$(MAKE) -C backend migrate
