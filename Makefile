include common.mk

log:
	git log --format="- %s" --reverse  | grep 'day-' > docs/log.txt
