#!/bin/sh

WOEID=2367105 # Boston

curl "http://weather.yahooapis.com/forecastrss?w=${WOEID}" > .cache/weather
