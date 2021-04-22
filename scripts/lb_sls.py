#!/usr/bin/python3
"""
    This script sends status of the lb service to SLS

"""
import argparse
import re
import sys
import time
import logging
import logging.config
import json
from datetime import datetime
import os
import requests



def get_arguments():
    """ Parse command line arguments"""
    parser = argparse.ArgumentParser(
        description='Gather heavy users of lxplus.')
    parser.add_argument('--debug', dest='debug', action='store_true',
                        help='write debug messages',
                        default=False)

    args = parser.parse_args()
    return args


def send(document):
    """ send the document """
    return requests.post('http://monit-metrics:10012/',
                         data=json.dumps(document),
                         headers={"Content-Type": "application/json; charset=UTF-8"})

def get_server_availability(logger, server_host):
    """ Contacts a server, and returns if it is up or not"""

    try:
        info = requests.get("http://%s/load-balancing/heartbeat" % server_host)
        logger.debug("Host contacted, and got %s", info.content.decode())
        my_date = int(re.match(r'.*: (\d+) : I am alive', info.content.decode()).group(1))

        logger.debug("The last execution was at %s", my_date)
        now = time.mktime(datetime.now().timetuple())
        latency = now - my_date
        if latency < 1800:
            logger.info("The server %s is up and running", server_host)
            return 100
        if latency < 7200:
            logger.warning("The server %s has not run for two hours", server_host)
            return 50
        logger.error("The server is down for more than two hours")
    except requests.exceptions.ConnectionError:
        logger.error("Error getting the info from %s", server_host)
    except AttributeError:
        logger.error("Error extracting the timestamp from the response" % info.content)


    return 0

def get_number_of_clusters(logger):
    """ Checks how many aliases are defined in ermis"""
    logger.info("Getting the number of aliases from kermis")

    return os.popen('/usr/bin/kermis -j -o read -a all | /usr/bin/jq ".[] |.alias_name "').read().count('\n')


def get_data(logger, args):
    """ Gets the KPI for the selected period"""
    logger.info("Ready to get the data for %s", args)

    availability = get_server_availability(logger, "lbmaster.cern.ch")

    if availability == 0:
        availability = get_server_availability(logger, "lbslave.cern.ch") / 2.

    number_of_clusters = get_number_of_clusters(logger)

    availabilitydesc = """<h3>DNS Load Balancing</h3><p>%s LB aliases defined</p>
<h4>Please follow the link below to see the LB Alias logs</h4><p>
<a href=\"https://aiermis.cern.ch/lbweb/logsform\">https://aiermis.cern.ch/lbweb/logsform</a></p>
""" % number_of_clusters

    sls_state = 'unavailable'

    if number_of_clusters == 0:
        sls_state = 'degraded'
        logger.error('Error getting the number of aliases from kermis')

    if availability > 75:
        sls_state = 'available'
    elif availability > 40:
        sls_state = 'degraded'


    return {'producer': 'loadbalancer',
            'type': 'availability',
            'serviceid': 'DNSLOADBALANCING',
            'service_status': sls_state,
            'timestamp': int(1000*time.mktime(datetime.now().timetuple())),
            'availabilityinfo': availabilitydesc,
            'availabilitydesc': "The availability has been estimated to %s" % availability,
            'contact': 'lb-experts@cern.ch',
            'webpage': 'http://information-technology.web.cern.ch/services/load-balancing-services',
           }

def send_and_check(logger, document):
    """ Sends a document to UMA"""
    response = requests.post('http://monit-metrics:10012/', data=json.dumps(document),
                             headers={"Content-Type": "application/json; charset=UTF-8"})
    logger.info("We got %s and %s ", response.status_code, response.text)
    assert(response.status_code in [200]),  \
           'With document: {0}. Status code: {1}. Message: {2}'.format(document,
                                                                       response.status_code,
                                                                       response.text)

def main():
    """ Let's do the the alarms"""
    args = get_arguments()
    logger = logging.getLogger(__name__)
    logger.propagate = False
    if args.debug:
        logger.setLevel(logging.DEBUG)
        formatter = logging.Formatter(
            '%(asctime)s - %(name)s %(funcName)20s() - %(levelname)s - %(message)s')
    else:
        logger.setLevel(logging.INFO)
        formatter = logging.Formatter(
            '%(asctime)s -  %(levelname)s - %(message)s')
    chan = logging.StreamHandler()
    chan.setFormatter(formatter)
    logger.addHandler(chan)

    document = get_data(logger, args)
    logger.info("The document is %s", document)
    send_and_check(logger, document)

    logger.info("Done")
    return 0


if __name__ == '__main__':
    sys.exit(main())
