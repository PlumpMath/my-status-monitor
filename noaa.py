#!/usr/bin/env python2

# Quick and dirty script to get the current temp from the National weather
# service.
#
# Dependencies:
#
# * requests -- for http requests
# * arrow -- date/time stuff
# * defusedxml -- for xml safety.
import requests
import arrow
import json
import sys
from defusedxml import ElementTree as ET

ENDPOINT = 'http://graphical.weather.gov/xml/sample_products/browser_interface/ndfdXMLclient.php'

ret = requests.get(ENDPOINT, params={
    'zipCodeList': '02215',
    'unit': 'e',  # Standard/English; 'm' denotes metric.
})

root = ET.fromstring(ret.text)


def build_time_layouts(tree):
    """Extracts a dict of time layouts from the element tree.

    Keys are the values of the 'layout-key' elements, values are lists
    of (start-valid-time, end-valid-time) pairs, each time being of type
    arrow, with end times possibly None.
    """
    layouts = {}
    for tl in tree.findall('.//time-layout'):
        # Find the layout key:
        key = tl.find('layout-key')
        if key is None:
            continue
        key = key.text

        # Convert list of start/end-valid-time elements to list of arrow
        def arrow_times(elements): return [arrow.get(e.text) for e in elements]
        starts = arrow_times(tl.findall('./start-valid-time'))
        ends = arrow_times(tl.findall('./end-valid-time'))
        if ends == []:
            ends = [None for s in starts]

        layouts[key] = zip(starts, ends)
    return layouts

layouts = build_time_layouts(root)


def get_temps(tree):
    elt = tree.find(".//temperature[@type='hourly']")
    temps = [float(e.text) for e in elt.findall('./value')]
    time_layout = [start for (start, end) in
                   layouts[elt.attrib['time-layout']]]
    if len(time_layout) != len(temps):
        print("length mismatch!")
    return zip(time_layout, temps)

temps = get_temps(root)

now = arrow.now()
current_temp = None

for (time, temp) in temps:
    if time > now:
        break
    current_temp = temp

if current_temp is None:
    sys.exit('No reading for right now')

print(json.dumps({"current_temp": current_temp, "units": "F"}))
