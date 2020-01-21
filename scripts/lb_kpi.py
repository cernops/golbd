#!/usr/bin/python
"""
    This script sends KPI of the lb service

"""
import argparse
import ConfigParser
import sys
import logging
import logging.config
import json
from datetime import datetime, timedelta
import os
import requests

from elasticsearch6 import Elasticsearch


def load_config():
    """  Load configuration  """
    primary_conf_path = "./lb.conf"
    secondary_conf_path = "/etc/cgs/cgs.conf"

    conf = ConfigParser.ConfigParser()

    if os.path.isfile(primary_conf_path):
        logging.config.fileConfig(
            primary_conf_path, disable_existing_loggers=False)
        logger = logging.getLogger(__name__)
        logger.info("Reading configuration from %s.", primary_conf_path)
        conf.read(primary_conf_path)
    elif os.path.isfile(secondary_conf_path):
        logging.config.fileConfig(
            secondary_conf_path, disable_existing_loggers=False)
        logger = logging.getLogger(__name__)
        logger.info("Reading configuration from %s.", secondary_conf_path)
        conf.read(secondary_conf_path)
    else:
        sys.stderr.write("No configuration file found! (Should be either %s or %s.)" %
                         (primary_conf_path, secondary_conf_path))
        sys.exit(1)

    return conf


def get_arguments():
    """ Parse command line arguments"""
    parser = argparse.ArgumentParser(
        description='Gather heavy users of lxplus.')
    parser.add_argument('--from', dest='start_date', default=None,
                        help='starting date')
    parser.add_argument('--to', dest='end_date', default=None,
                        help='end date')
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


def get_data(logger, args):
    """ Gets the KPI for the selected period"""
    logger.info("Ready to get the data for %s", args)

    my_conf = load_config()

    username = my_conf.get("elasticsearch", "user_lb")
    password = my_conf.get("elasticsearch", "password_lb")

    epoch = int(1000*(datetime.utcnow() - datetime(1970, 1, 1)).total_seconds())
    my_data = []

    try:
        my_es = Elasticsearch([{"host": "es-timber.cern.ch",
                                "port": 443, "url_prefix": "es",
                                "http_auth": username + ":" + password}],
                              use_ssl=True, verify_certs=True,
                              ca_certs="/etc/pki/tls/certs/ca-bundle.trust.crt")
        result = my_es.search("monit_prod_loadbalancer_logs_ermis_latest",
                              body={"size": 0,
                                    "aggs": {"tenant": {
                                        "terms": {"field": "data.tenant"},
                                        "aggs": {"clusters":
                                                 {"cardinality": {
                                                     "field": "data.cluster"}},
                                                 "cnames":
                                                 {"cardinality": {
                                                     "field": "data.cname_record"}}}}}})
        logger.info(result['aggregations'])
        tenants = {}
        for data in result['aggregations']['tenant']['buckets']:
            logger.info(data)
            tenants[data['key']] = {
                'number_of_clusters': data['clusters']['value'],
                'number_of_cnames': data['cnames']['value'],
                'number_of_nodes': 0
            }
        result = my_es.search("monit_prod_loadbalancer_logs_server*",
                              body={"size": 0,
                                    "query": {"range": {"metadata.timestamp": {"gte": "now-2h"}}},
                                    "aggs": {"tenant": {
                                        "terms": {"field": "data.partition_name"},
                                        "aggs": {"nodes": {"cardinality": {"field": "data.node"}}}
                                    }}})

        for data in result['aggregations']['tenant']['buckets']:
            logger.info(data)
            if not tenants.has_key(data['key']):
                logger.info("Skipping %s", data['key'])
                continue
            tenants[data['key']]['number_of_nodes'] = data['nodes']['value']

        for tenant in tenants:
            logger.info(tenant)

            my_data.append({'timestamp': epoch,
                            'producer': 'loadbalancer',
                            'idb_fields': ["number_of_clusters", "number_of_cnames",
                                           "number_of_nodes"],
                            'idb_tags': ['partition'],
                            'type': 'kpi',
                            'number_of_clusters': tenants[tenant]['number_of_clusters'],
                            'number_of_nodes': tenants[tenant]['number_of_nodes'],
                            'number_of_cnames': tenants[tenant]['number_of_cnames'],
                            'partition': tenant})

    except KeyError, my_ex:
        logger.error("Error connecting to elasticsearch: %s", my_ex)

    return my_data


def send_kpi(logger, data):
    """ Get the kpi values and sends them to MONIT"""

    logger.info("Ready to send %s", data)
    response = send(data)
    logger.info("Sent the message with %i", response.status_code)
    if response.status_code != 200:
        logger.error("Error sending the kpi to monit")
        return False
    return True


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

    data = get_data(logger, args)

    send_kpi(logger, data)

    logger.info("Done")
    return 0


if __name__ == '__main__':
    sys.exit(main())
