# Copyright (c) 2026, WSO2 LLC. (http://www.wso2.com).
# Licensed under the Apache License, Version 2.0.
"""Run branch migration fixtures through the existing Nutcracker corpus harness."""
import concurrent.futures
import json
import os
from pathlib import Path
import re
import signal
import shutil
import tomllib
import subprocess
import time

ROOT = Path('/workspace/scratch/jballerina-migration/migration-results')
OUT = Path('/workspace/ballerina-lang-go/triage/migrated-tests')
SUFFIX = re.compile(r'-(?:[vep]|f[vep])(?:\.bal)?$')
JAVA = re.compile(r'import\s+(?:ballerina/jballerina\.java|ballerinax/java)(?:\s|;|\.)|@java\s*:|platform\.java\d*|\.jar["\']', re.I)

def prepare_workspace(case, files):
    if not case['Project']:
        return files
    path = Path(case['Path'])
    workspace = next((parent for parent in path.parents if parent != ROOT and ROOT in parent.parents and (parent/'Ballerina.toml').exists() and '[workspace]' in (parent/'Ballerina.toml').read_text()), None)
    if workspace is None:
        return files
    members = tomllib.loads((workspace/'Ballerina.toml').read_text())['workspace']['packages']
    target = str(path.relative_to(workspace))
    descriptors = {}
    for member in members:
        pkg = tomllib.loads((workspace/member/'Ballerina.toml').read_text())['package']
        descriptors[(pkg['org'],pkg['name'])] = member
    selected = {target}
    todo = [target]
    while todo:
        member = todo.pop()
        for source in (workspace/member).rglob('*.bal'):
            for org, name in re.findall(r'(?m)^\s*import\s+([\w]+)/([\w]+)', source.read_text()):
                dependency = descriptors.get((org,name))
                if dependency is not None and dependency not in selected:
                    selected.add(dependency)
                    todo.append(dependency)
    ordered = [target, *sorted(selected-{target})]
    dest = Path('/workspace/scratch/nutcracker-triage-workspaces')/case['Id']
    if dest.exists():
        shutil.rmtree(dest)
    dest.mkdir(parents=True,exist_ok=True)
    for member in ordered:
        shutil.copytree(workspace/member, dest/member, dirs_exist_ok=True, ignore=shutil.ignore_patterns('target','build','.git'))
    (dest/'Ballerina.toml').write_text('[workspace]\npackages = '+json.dumps(ordered)+'\n')
    case['WorkspaceRoot'] = str(workspace)
    case['WorkspaceMembers'] = ordered
    case['ExecutionPath'] = str(dest)
    return [workspace/'Ballerina.toml', *(p for member in ordered for p in (workspace/member).rglob('*') if p.is_file())]

def discover():
    cases = []
    projects = sorted(p.parent for p in ROOT.rglob('Ballerina.toml') if '[package]' in p.read_text())
    for project in projects:
        if SUFFIX.search(project.name):
            name = str(project.relative_to(ROOT))
        else:
            mains = [p for p in project.glob('*.bal') if SUFFIX.search(p.name)]
            if len(mains) != 1:
                raise ValueError(f'Ambiguous project {project}: {mains}')
            name = str(mains[0].relative_to(ROOT))
        cases.append({'Name':name,'Path':str(project),'Project':True})
    for source in sorted(ROOT.rglob('*.bal')):
        if any(p == source.parent or p in source.parents for p in projects):
            continue
        if not SUFFIX.search(source.name):
            raise ValueError(f'Unclassified standalone source: {source}')
        cases.append({'Name':str(source.relative_to(ROOT)),'Path':str(source),'Project':False})
    for i, case in enumerate(sorted(cases, key=lambda c:c['Name'])):
        case['Id'] = f'T{i+1:04d}'
        path = Path(case['Path'])
        files = sorted(path.rglob('*')) if case['Project'] else [path]
        # Include workspace member dependencies and the manifest which governs standalone files.
        for ancestor in [path.parent, *path.parents]:
            if ancestor == ROOT or ROOT not in ancestor.parents:
                break
            manifest = ancestor/'Ballerina.toml'
            if manifest.exists():
                if '[workspace]' in manifest.read_text():
                    files += sorted(ancestor.rglob('*'))
                else:
                    files.append(manifest)
                break
        files = sorted(set(p for p in prepare_workspace(case, files) if p.is_file()))
        evidence = []
        for file in files:
            if file.suffix in ('.bal','.toml'):
                for number,line in enumerate(file.read_text().splitlines(),1):
                    if JAVA.search(line):
                        evidence.append(f'{file.relative_to(ROOT)}:{number}: {line.strip()}')
        case['JavaEvidence'] = evidence
        case['Sources'] = [str(p) for p in files if p.suffix in ('.bal','.toml')]
        args = []
        for p in files:
            if p.suffix == '.bal':
                args += re.findall(r'^\s*//\s*@args\s+(\[.*\])\s*$', p.read_text(), re.M)
        if len(args)>1:
            case['ArgsProblem'] = 'Multiple @args annotations'
        case['Args'] = json.loads(args[0]) if len(args)==1 else None
        yield case

def execute(case):
    folder = OUT/'evidence'/case['Id']
    folder.mkdir(parents=True,exist_ok=True)
    if case['JavaEvidence']:
        return dict(case, Status='excluded-java')
    env = os.environ.copy()
    env.update(MIGRATION_TRIAGE_CASE=json.dumps(dict(case, Path=case.get('ExecutionPath',case['Path']))), MIGRATION_TRIAGE_CAPTURE=str(folder/'capture.json'), MIGRATION_TRIAGE_BAL='/workspace/ballerina-lang-go/bal', GOMAXPROCS='2')
    started=time.monotonic()
    proc=subprocess.Popen(['/workspace/scratch/migration-triage.test','-test.run=^TestMigrationTriage$','-test.v','-test.timeout=25s'],cwd=ROOT,env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,start_new_session=True)
    timeout=False
    try:
        log,_=proc.communicate(timeout=30)
    except subprocess.TimeoutExpired:
        timeout=True
        os.killpg(proc.pid,signal.SIGKILL)
        log,_=proc.communicate()
    (folder/'run.log').write_bytes(log)
    result=dict(case,Status='pass' if proc.returncode==0 else ('timeout' if timeout or b'test timed out' in log else 'fail'), ExitCode=proc.returncode, Seconds=round(time.monotonic()-started,3))
    (folder/'result.json').write_text(json.dumps(result,indent=2)+'\n')
    return result

if __name__=='__main__':
    cases=list(discover())
    (OUT/'inventory.json').write_text(json.dumps(cases,indent=2)+'\n')
    print(f'Inventory: {len(cases)} cases; {sum(bool(c["JavaEvidence"]) for c in cases)} Java exclusions',flush=True)
    with concurrent.futures.ThreadPoolExecutor(max_workers=6) as pool, (OUT/'results.jsonl').open('w') as stream:
        for count,result in enumerate(pool.map(execute,cases),1):
            stream.write(json.dumps(result)+'\n');stream.flush()
            if count%50==0: print(f'Completed {count}/{len(cases)}',flush=True)
