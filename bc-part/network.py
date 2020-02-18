#!/usr/bin/python3

import argparse
import os
import time
from pathlib import Path
from shutil import copyfile

import yaml
from mako.template import Template

with open("hosts.yml", 'r') as stream:
    try:
        config = yaml.safe_load(stream)
        config = config['all']
    except Exception as ex:
        print(ex)
        exit(1)

DOMAIN = config['vars']['global_domain']

ORG_VARS = {'name': 'org', 'www_port': 'www_port', 'ca_port': 'ca_port', 'couchdb_port': 'couchdb_port',
            'org_ou': 'org_ou', 'node_roles': 'node_roles'}
ORGS = [{**org, 'n': num, 'PEER0_PORT': 7051 + 1000 * num, 'PEER0_EVENT_PORT': 7053 + 1000 * num} for num, org in
        enumerate(
            [{k_to: v[k_from] for (k_to, k_from) in ORG_VARS.items()} for v in
             config['children']['nodes']['hosts'].values()])]
ORG_NAMES = list(map(lambda org: org['name'], ORGS))

TMPL_ARTIFACTS_DIR = config['generate_vars']['TMPL_ARTIFACTS_DIR']
TMPL_DOCKER_COMPOSE_DIR = config['generate_vars']['TMPL_DOCKER_COMPOSE_DIR']
ARTIFACTS_DIR = config['generate_vars']['ARTIFACTS_DIR']
DOCKER_COMPOSE_DIR = config['generate_vars']['DOCKER_COMPOSE_DIR']
FABRIC_VERSION = config['generate_vars']['FABRIC_VERSION']
REST_API_IMAGE = config['generate_vars']['REST_API_IMAGE']
NETWORK = config['generate_vars']['NETWORK']

EXPLORER_PORT = config['vars']['default_hl_explorer_port']
EXPLORER_USER = config['vars']['default_hl_explorer_username']
EXPLORER_PASSWORD = config['vars']['default_hl_explorer_password']

CHANNELS = config['vars']['global_channels']
channel = config['vars']['global_channels'][0]

CHANNEL = channel['name']
CHAINCODE_VERSION = channel['chaincode']['version']
CHAINCODE_DELARUE = channel['chaincode']['name']

CHAINCODE_DELARUE_POLICY = f'OR(' + (', '.join(map(lambda org: f'"{org["name"]}MSP.peer"', ORGS))) + ')'

UID = os.getuid()
GID = os.getgid()

os.environ["COMPOSE_PROJECT_NAME"] = NETWORK

IS_LOCAL = False


def copy_tmpl(src, target, dictionary):
    print(f'Process template {src}, creating {target}')
    with open(src) as file_src:
        with open(target, 'w') as file_target:
            file_target.write(Template(file_src.read()).render(**dictionary))


def run(cmd):
    print(cmd)
    os.system(cmd)


class BgColors:
    ENDC = '\033[0m'
    BOLD = '\033[1m'


def log(text, n=0):
    print(f'{BgColors.BOLD} -=- {text} -=- {BgColors.ENDC}' + '\n' * n)


def remove_generated():
    print(f'Removing generated and downloaded artifacts from: {DOCKER_COMPOSE_DIR}, {ARTIFACTS_DIR}')
    os.system(f'rm -rf {DOCKER_COMPOSE_DIR}')
    os.system(f'rm -rf {ARTIFACTS_DIR}')


def clean_docker():
    log("Remove docker containers, volumes and images")
    docker_ids_raw = os.popen(f"docker ps -a | grep '\.{DOMAIN}' | awk '{{print $1}}'").read()
    docker_ids = list(filter(None, docker_ids_raw.split('\n')))
    if docker_ids:
        log(f"Removing docker instances found with {DOMAIN}: {', '.join(docker_ids)}")
        os.system(f"docker rm -f {' '.join(docker_ids)}")
    else:
        log(f"No docker instances available for deletion with {DOMAIN}")

    docker_volumes_raw = os.popen(f"docker volume ls -q | grep '\.{DOMAIN}' | awk '{{print $1}}'").read()
    docker_volumes = list(filter(None, docker_volumes_raw.split('\n')))
    if docker_volumes:
        log(f"Removing docker volumes found with {DOMAIN}: {', '.join(docker_volumes)}")
        run(f"docker volume rm -f {' '.join(docker_volumes)}")
    else:
        log(f"No docker volumes available for deletion with {DOMAIN}")

    docker_images_raw = os.popen(f"docker image ls | grep '\.{DOMAIN}' | awk '{{print $1}}'").read()
    docker_images = list(filter(None, docker_images_raw.split('\n')))
    if docker_images:
        log(f"Removing docker images found with {DOMAIN}: {', '.join(docker_images)}")
        run(f"docker image rm -f {' '.join(docker_images)}")
    else:
        log(f"No docker images available for deletion with {DOMAIN}")


def generate_peer_artifacts(org):
    org_name = org['name']
    org_artifacts_dir = f'{ARTIFACTS_DIR}/{org_name}'
    org_compose_dir = f'{DOCKER_COMPOSE_DIR}/{org_name}'

    log(f'Generating artifacts for {org_name} into {org_artifacts_dir}')
    log(f'Generating docker compose files for {org_name} into {org_compose_dir}')

    os.makedirs(org_artifacts_dir, exist_ok=True)
    os.makedirs(org_compose_dir, exist_ok=True)

    # Universal dictionary for templates
    d = {'DOMAIN': DOMAIN, 'ORG': org, 'ORG_NAME': org_name, 'ORG_NAMES': ORG_NAMES, 'FABRIC_VERSION': FABRIC_VERSION,
         'REST_API_IMAGE': REST_API_IMAGE, 'NETWORK': NETWORK, 'CHANNEL': CHANNEL, 'CHAINCODE': CHAINCODE_DELARUE}

    if 'explorer' in org['node_roles']:
        d['EXPLORER_PORT'] = EXPLORER_PORT

    # Copy nginx proxy config
    copy_tmpl(f'{TMPL_ARTIFACTS_DIR}/nginx.conf', f'{org_artifacts_dir}/nginx.conf', d)

    # API configs
    log('Creating /api-configs folder')
    org_api_config_dir = f'{org_artifacts_dir}/api-configs'
    os.makedirs(org_api_config_dir, exist_ok=True)

    copy_tmpl(f'{TMPL_ARTIFACTS_DIR}/api-configs/api.yaml', f'{org_api_config_dir}/api.yaml', d)
    copy_tmpl(f'{TMPL_ARTIFACTS_DIR}/api-configs/network.yaml', f'{org_api_config_dir}/network.yaml', d)

    log('Creating cryptogen.yaml')
    copy_tmpl(f'{TMPL_ARTIFACTS_DIR}/cryptogen-peer.yaml', f'{org_artifacts_dir}/cryptogen.yaml', d)

    # Copy CA config
    copy_tmpl(f'{TMPL_ARTIFACTS_DIR}/fabric-ca-server-config.yaml',
              f'{org_artifacts_dir}/fabric-ca-server-config.yaml', d)

    # Copy docker-compose yaml
    copy_tmpl(f'{TMPL_DOCKER_COMPOSE_DIR}/peer.yaml', f'{org_compose_dir}/peer.yaml', d)

    if IS_LOCAL:
        copy_tmpl(f'{TMPL_DOCKER_COMPOSE_DIR}/local-peer.yaml', f'{org_compose_dir}/local-peer.yaml', d)

    log("Generating crypto material with cryptogen in domain CLI container")
    compose_utils = f'{DOCKER_COMPOSE_DIR}/cli.yaml'
    cli = f'cli.{DOMAIN}'
    run(f'docker-compose --file {compose_utils} up -d {cli}')

    log("Generate crypto-config inside CLI")
    run(
        f'docker exec {cli} bash -c "cryptogen generate --config={org_name}/cryptogen.yaml"')

    log("Changing ownership")
    run(f'docker exec {cli} bash -c "chown -R {UID}:{GID} ."')

    log("Copy CA private key to server.key")
    for filename in Path(ARTIFACTS_DIR).glob('**/ca/*_sk*'):
        ca_path = os.path.dirname(os.path.abspath(filename))
        copyfile(os.path.abspath(filename), f'{ca_path}/server.key')

    log(f"Removing CLI image - {cli}")
    run(f'docker stop {cli} && docker rm {cli}')

    log(f'Done generating artifacts for {org_name}', 1)


def generate_orderer_docker_compose(org):
    org_compose_dir = f'{DOCKER_COMPOSE_DIR}/{org["name"]}'

    log(f"Creating orderer docker-compose file for {org['name']}")

    compose_template = f'{TMPL_DOCKER_COMPOSE_DIR}/orderer.yaml'
    compose_out = f"{org_compose_dir}/orderer.yaml"

    d = {'ORG_N': org['n'], 'DOMAIN': DOMAIN, 'NETWORK': NETWORK, 'FABRIC_VERSION': FABRIC_VERSION}

    copy_tmpl(compose_template, compose_out, d)


def generate_domain_docker_compose():
    log(f"Creating domain docker-compose file")

    d = {'DOMAIN': DOMAIN, 'NETWORK': NETWORK, 'FABRIC_VERSION': FABRIC_VERSION}

    copy_tmpl(src=TMPL_DOCKER_COMPOSE_DIR + '/base.yaml', target=DOCKER_COMPOSE_DIR + '/base.yaml', dictionary=d)
    copy_tmpl(f'{TMPL_DOCKER_COMPOSE_DIR}/domain-cli.yaml', f"{DOCKER_COMPOSE_DIR}/cli.yaml", d)


def generate_channel_artifacts():
    log("Creating channel artifacts")

    os.makedirs(f"{ARTIFACTS_DIR}/channel", exist_ok=True)

    d = {'ORGS': ORGS, 'ORG_NAMES': ORG_NAMES, 'DOMAIN': DOMAIN, 'ORGS_COUNT': len(ORGS)}

    create_channels = [CHANNEL]
    compose_domain_cli = f"{DOCKER_COMPOSE_DIR}/cli.yaml"
    cli = f'cli.{DOMAIN}'

    run(f'docker-compose --file {compose_domain_cli} up -d {cli}')

    copy_tmpl(f"{TMPL_ARTIFACTS_DIR}/configtx-template.yaml", f'{ARTIFACTS_DIR}/configtx.yaml', d)
    copy_tmpl(f'{TMPL_ARTIFACTS_DIR}/cryptogen-orderer.yaml', f'{ARTIFACTS_DIR}/cryptogen-{DOMAIN}.yaml', d)

    log("Generating crypto material with cryptogen")

    run(f'docker exec "{cli}" bash -c "cryptogen generate --config=cryptogen-{DOMAIN}.yaml"')

    log("Generating genesis block")
    run(f'docker exec -e FABRIC_CFG_PATH=/etc/hyperledger/artifacts {cli} '
        f'configtxgen -profile OrdererGenesis -outputBlock ./channel/genesis.block')

    for channel_name in create_channels:
        log(f"Generating channel config transaction for {channel_name}")
        run(
            f'docker exec -e FABRIC_CFG_PATH=/etc/hyperledger/artifacts {cli} '
            f'configtxgen -profile "{channel_name}" -outputCreateChannelTx "./channel/{channel_name}.tx" -channelID "{channel_name}"')

        for myorg in ORG_NAMES:
            log(f"Generating anchor peers update for {myorg}")
            run(f'docker exec -e FABRIC_CFG_PATH=/etc/hyperledger/artifacts {cli} '
                f'configtxgen -profile "{channel_name}" '
                f'--outputAnchorPeersUpdate "./channel/{myorg}MSPanchors-{channel_name}.tx" '
                f'-channelID "{channel_name}" -asOrg {myorg}MSP')

    log("Changing ownership")
    run(f'docker exec {cli} bash -c "chown -R {UID}:{GID} ."')

    run(f'docker stop {cli} && docker rm {cli}')

def generate_explorer_config(org):
    log("Generate explorer config")
    org_name = org['name']
    org_artifacts_dir = f'{ARTIFACTS_DIR}/{org_name}'

    d = {
        'DOMAIN': DOMAIN,
        'ORG': org,
        'ORG_NAME': org_name,
        'CHANNELS': CHANNELS,
        'EXPLORER_USER': EXPLORER_USER,
        'EXPLORER_PASSWORD': EXPLORER_PASSWORD,
    }
    copy_tmpl(f'{TMPL_ARTIFACTS_DIR}/explorer-config.json', f'{org_artifacts_dir}/explorer-config.json', d)
    run(f"chmod 755 {org_artifacts_dir}/explorer-config.json")

def copy_sk_files():
    for root, dir, files in os.walk(f'{ARTIFACTS_DIR}'):
        for file in files:
            if (file.endswith("_sk")):
                os.system(f'cp {root + os.sep + file} {root + os.sep + "server.key"}')    


def generate():
    clean_docker()
    remove_generated()

    for folder in [ARTIFACTS_DIR, DOCKER_COMPOSE_DIR]:
        os.makedirs(folder, exist_ok=True)

    generate_domain_docker_compose()

    for org in ORGS:
        generate_peer_artifacts(org)
        generate_orderer_docker_compose(org)
        if 'explorer' in org['node_roles']:
            generate_explorer_config(org)

    generate_channel_artifacts()
    copy_sk_files()


def docker_compose_up(org_name):
    compose_file = f"{DOCKER_COMPOSE_DIR}/{org_name}/peer.yaml"
    if IS_LOCAL:
        compose_file += f" -f {DOCKER_COMPOSE_DIR}/{org_name}/local-peer.yaml"
    log(f"Starting docker instances from {compose_file}")
    run(f'docker-compose -f {compose_file} up -d')

    compose_file = f"{DOCKER_COMPOSE_DIR}/{org_name}/orderer.yaml"
    log(f"Starting docker instances from {compose_file}")
    run(f'docker-compose -f {compose_file} up -d')


def install_chaincode(org_name, cc_name, cc_version):
    lang = 'golang'

    log(f"Installing chaincode {cc_name} to peers of {org_name} from ./chaincode/go/{cc_name} version {cc_version}")
    run(
        f'docker exec "cli.{org_name}.{DOMAIN}" bash -c "CORE_PEER_ADDRESS=peer0.{org_name}.{DOMAIN}:7051 peer chaincode install -n {cc_name} -v {cc_version} -p {cc_name} -l {lang}"')


def install_all(org_name):
    for cc_name in [CHAINCODE_DELARUE]:
        install_chaincode(org_name, cc_name, CHAINCODE_VERSION)


def create_channel(org_name, channel_name):
    cli = f'cli.{org_name}.{DOMAIN}'

    log(f"Creating channel {channel_name} by {org_name} using {cli}")

    run(
        f'docker exec {cli} bash -c "peer channel create -o orderer0.{DOMAIN}:7050 -c {channel_name} '
        f'-f /etc/hyperledger/artifacts/channel/{channel_name}.tx '
        f'--tls --cafile /etc/hyperledger/crypto/orderer/tls/ca.crt"')

    run(
        f'docker exec {cli} bash -c "peer channel update -o orderer0.{DOMAIN}:7050 -c {channel_name} '
        f'-f /etc/hyperledger/artifacts/channel/{org_name}MSPanchors-{channel_name}.tx '
        f'--tls --cafile /etc/hyperledger/crypto/orderer/tls/ca.crt"')

    log("Changing ownership of channel block files")
    run(f'docker exec {cli} bash -c "chown -R {UID}:{GID} ."')


def join_channel(org_name, channel_name):
    log(f"Joining channel {channel_name} by all peers of {org_name}")
    run(
        f'docker exec "cli.{org_name}.{DOMAIN}" bash -c "CORE_PEER_ADDRESS=peer0.{org_name}.{DOMAIN}:7051 peer channel join -b {channel_name}.block"')


def instantiate_chaincode(org_name, channel_name, cc_name, cc_init, policy):
    log(f"Instantiating chaincode {cc_name} on {channel_name} by {org_name}")
    c = f"CORE_PEER_ADDRESS=peer0.{org_name}.{DOMAIN}:7051 peer chaincode instantiate -n {cc_name} -v {CHAINCODE_VERSION} -P '{policy}' -c '{cc_init}' -o orderer0.{DOMAIN}:7050 -C {channel_name} --tls --cafile /etc/hyperledger/crypto/orderer/tls/ca.crt"
    c = c.replace('"', r'\"')

    log(f"Instantiating {cc_name} for {org_name}")
    run(f'docker exec cli.{org_name}.{DOMAIN} bash -c "{c}"')


def create_join_instantiate(org_name, channel_name, cc_name, cc_init, policy):
    create_channel(org_name, channel_name)
    join_channel(org_name, channel_name)
    instantiate_chaincode(org_name, channel_name, cc_name, cc_init, policy)


def up():
    log("Up")

    for org_name in ORG_NAMES:
        docker_compose_up(org_name)

    # TODO wait for raft leader election, for now just wait for 15 seconds
    time.sleep(15)

    for org_name in ORG_NAMES:
        install_all(org_name)

    create_join_instantiate(ORGS[0]['name'], CHANNEL, CHAINCODE_DELARUE, '{"Args":[]}', CHAINCODE_DELARUE_POLICY)

    for org_name in ORG_NAMES[1:]:
        join_channel(org_name, CHANNEL)

    #Restart explorer and set cron task
    pwd = os.path.dirname(os.path.abspath(__file__))
    for org in ORGS:
        if 'explorer' in org['node_roles']:
            run(f"docker exec explorer-db.{org['name']}.{DOMAIN} /bin/bash /opt/createdb.sh")
            time.sleep(15)
            run(f"docker restart explorer.{org['name']}.{DOMAIN}")
            run(f"echo '0 */3 * * * root docker rm -f explorer.{org['name']}.{DOMAIN} || cd {pwd} && docker-compose --file docker-compose/{org['name']}/peer.yaml up -d explorer.{org['name']}.{DOMAIN} 2>&1' | sudo tee /etc/cron.d/explorer-{org['name']}")


parser = argparse.ArgumentParser()
parser.add_argument("action", help="Perform Fabric network action")
parser.add_argument("--local", help="Local deployment", action="store_true")
args = parser.parse_args()

if args.local:
    IS_LOCAL = True

if args.action == 'up':
    up()
elif args.action == 'generate':
    generate()
elif args.action == 'clean':
    remove_generated()
    clean_docker()
else:
    log("Action not found")
