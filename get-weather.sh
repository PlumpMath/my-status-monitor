#!/bin/sh

WOEID=2367105 # Boston

curl -f "https://weather.yahooapis.com/forecastrss?w=${WOEID}" > $HOME/.cache/weather
