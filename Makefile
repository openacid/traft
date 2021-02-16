include common.mk

log:
	git log --format="- %ai %s" --reverse  | grep 'day-' > docs/log.txt
