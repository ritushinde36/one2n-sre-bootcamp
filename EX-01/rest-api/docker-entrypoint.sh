#!/bin/sh
set -e

migrate up
exec rest-api
