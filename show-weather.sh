#!/bin/sh

cat $HOME/.cache/weather | xml2 | awk -F= '

/yweather:units\/@temperature/ { tempunit = $2 }
/yweather:condition\/@text/ { text = $2 }
/yweather:condition\/@temp/ { temp = $2 }
/yweather:wind\/@chill/ { windchill = $2 }
/yweather:atmosphere\/@humidity/ { humidity= $2 }

END {
	#print text ", " temp " " tempunit " (Humidity: " humidity")"
	print text ", " temp " " tempunit " (Wind Chill: " windchill")"
}

'
