#!/usr/bin/env python

from docker import Client
import os
import requests
import socket
import time

import print_utils

HTTP_LOCALHOST = "http://localhost"

cli = Client(base_url="unix://var/run/docker.sock")
max_retry_attempts = 100
default_host = "localhost"
healthcheck_path = "/health"
sleep_time_secs = 1


#
# Get container local ip.
# IP address returned is only reachable on the local machine and within
# the container.
#
def get_container_ip(container_name):
    cmd = "docker inspect "
    cmd += '-f "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}" %s'
    return os.popen(cmd % container_name).read().strip()


#
# Returns whether the zk listening on the given port is ready.
#
def is_zk_ready(port):
    cmd = "bash -c 'echo ruok | nc localhost {}'".format(port)
    return os.popen(cmd).read().strip() == "imok"


#
# Returns a free port on the host.
#
def find_free_port():
    s = socket()
    s.bind(('', 0))
    port = s.getsockname()[1]
    s.close()
    return port


#
# Force remove container by name (best effort)
#
def remove_existing_container(name):
    try:
        cli.remove_container(name, force=True)
        print_utils.okblue("removed container %s" % name)
    except Exception as e:
        if "No such container" in str(e):
            return
        raise e


#
# Stop container by name
#
def stop_container(name):
    try:
        cli.stop(name, timeout=5)
        print_utils.okblue("stopped container %s" % name)
    except Exception as e:
        if "No such container" in str(e):
            return
        raise e


#
# Run health check for peloton apps
#
def wait_for_up(app, port, path=healthcheck_path):
    count = 0
    error = ""
    url = "http://%s:%s/%s" % (default_host, port, path)
    while count < max_retry_attempts:
        try:
            r = requests.get(url)
            if r.status_code == 200:
                print_utils.okgreen("started %s" % app)
                return
        except Exception as e:
            if count % 5 == 1:
                msg = "app %s is not up yet, retrying at %s" % (app, url)
                print_utils.warn(msg)
            error = str(e)
            time.sleep(sleep_time_secs)
            count += 1

    raise Exception(
        "failed to start %s on %d after %d attempts, err: %s"
        % (app, port, max_retry_attempts, error)
    )
