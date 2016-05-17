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
ZIPCODE = '02215'

layouts = {}


class LengthMismatch(Exception):

    def __init__(self, l, r):
        Exception.__init__(self, 'Length mismatch: len(%r) != len(%r)' %
                           (l, r))


def fetch_data():
    """Fetch the data from NOAA.

    Returns an element tree of the parsed xml.
    """
    resp = requests.get(ENDPOINT, params={
        'zipCodeList': ZIPCODE,
        'unit': 'e',  # Standard/English; 'm' denotes metric.
    })
    return ET.fromstring(resp.text)


def build_time_layouts(tree):
    """Extracts a dict of time layouts from the element tree.

    Keys are the values of the 'layout-key' elements, values are lists
    of (start-valid-time, end-valid-time) pairs, each time being of type
    arrow, with end times possibly None.
    """
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


def time_map_for(tree, parent_query, child_query):
    """Return a list of pairs mapping times to elements.

    - `tree` should be the root of the element tree.
    - `parent_query` must be an xpath query, which when executed on `tree`
       will select an element with a 'time-layout' attribute.
    - `child_query` must be an xpath query, which when executed on the parent
      will select one element for each entry in the time layout.

    The times will be taken from the 'time-layout' parameter of the parent
    element, and the values will be the child elements.
    """
    parent = tree.find(parent_query)
    time_layout = layouts[parent.attrib['time-layout']]
    children = parent.findall(child_query)
    if len(time_layout) != len(children):
        raise LengthMismatch(time_layout, children)
    return zip(time_layout, children)


def entry_for_time(time_map, time):
    """Return the first entry in `time_map` whose time range contains `time`.

    `time_map` should be a value returend by `time_map_for`. `time` should be
    an arrow.

    Return none if no such range exists.
    """
    if len(time_map) == 0:
        return None
    result = time_map[0][1]
    for ((start, end), value) in time_map:
        if start <= time and (end is None or end > time):
            result = value
    return result


def get_hourly_temps(tree):
    """Return a mapping from an hourly time layout to temperatures."""
    time_map = time_map_for(tree, ".//temperature[@type='hourly']", './value')
    return [(k, float(v.text)) for k, v in time_map]


def get_hourly_conditions(tree):
    """Return a mapping from a time layout to conditions."""
    time_map = time_map_for(tree, './/weather', './weather-conditions')
    time_map = [(k, v.find('./value')) for (k, v) in time_map]

    last = None
    for i in range(len(time_map)):
        k, v = time_map[i]
        if v is None:
            v = last
        else:
            last = v
        time_map[i] = (k, v)

    for i in range(len(time_map)):
        k, v = time_map[i]
        if v is not None:
            v = dict([(attr, v.attrib[attr])
                    for attr in ('coverage',
                                'intensity',
                                'weather-type')])
        time_map[i] = (k, v)
    return time_map


if __name__ == '__main__':
    root = fetch_data()
    build_time_layouts(root)

    now = arrow.now()
    temps = get_hourly_temps(root)
    conditions = get_hourly_conditions(root)

    current_temp = entry_for_time(temps, now)

    if current_temp is None:
        sys.exit('No reading for right now')

    print(json.dumps({
        "temp": current_temp,
        "units": "F",
        "conditions": entry_for_time(conditions, now),
    }))
