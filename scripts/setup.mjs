#!/usr/bin/env node
/**
 * Yomiage Keiryou cross-platform setup orchestrator.
 *
 * This file intentionally invokes executables directly without shell syntax so
 * that the same flow works on Windows, Linux, and WSL.
 */
import { existsSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const DOCKER_DIR = resolve(ROOT, 'windows', 'docker');
const isWindows = process.platform === 'win32';

function usage() {
  console.log(`🗣️ Yomiage Keiryou 構築自動化

使い方:
  node scripts/setup.mjs [options]

オプション:
  --mode docker       Docker Desktop / Linux Dockerで構築（既定）
  --mode build        Linux amd64向けDAVE対応バイナリをビルド
  --mode verify       構成確認、シェル構文、ドライランを実行
  --skip-env          対話型.env設定を省略（既存.envが必要）
  --force-env         .envを再設定し、利用ルールへ再同意
  --no-build          Dockerイメージの再ビルドを省略
  --dry-run           実行せず、実行予定コマンドだけ表示
  --help              このヘルプを表示

例:
  node scripts/setup.mjs
  node scripts/setup.mjs --mode docker --force-env
  node scripts/setup.mjs --mode build
  node scripts/setup.mjs --mode verify
`);
}

function parseArgs(argv) {
  const options = {
    mode: 'docker',
    skipEnv: false,
    forceEnv: false,
    noBuild: false,
    dryRun: false,
  };
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === '--help' || arg === '-h') options.help = true;
    else if (arg === '--skip-env') options.skipEnv = true;
    else if (arg === '--force-env') options.forceEnv = true;
    else if (arg === '--no-build') options.noBuild = true;
    else if (arg === '--dry-run') options.dryRun = true;
    else if (arg === '--mode') {
      options.mode = argv[++i];
      if (!['docker', 'build', 'verify'].includes(options.mode)) {
        throw new Error(`未対応のmodeです: ${options.mode}`);
      }
    } else {
      throw new Error(`不明なオプションです: ${arg}`);
    }
  }
  return options;
}

function commandLabel(command, args) {
  return [command, ...args].map((value) => JSON.stringify(value)).join(' ');
}

function run(command, args, { env = {}, cwd = ROOT } = {}) {
  console.log(`\n▶ ${commandLabel(command, args)}`);
  if (currentOptions.dryRun) return;
  const result = spawnSync(command, args, {
    cwd,
    env: { ...process.env, ...env },
    stdio: 'inherit',
    shell: false,
    windowsHide: false,
  });
  if (result.error) throw result.error;
  if (result.status !== 0) {
    throw new Error(`コマンドが失敗しました（終了コード: ${result.status ?? '不明'}）`);
  }
}

function hasCommand(command, args = ['--version']) {
  const result = spawnSync(command, args, {
    cwd: ROOT,
    stdio: 'ignore',
    shell: false,
    windowsHide: true,
  });
  return !result.error && result.status === 0;
}

function requireCommand(command, label) {
  if (currentOptions.dryRun) {
    console.log(`ℹ️ [dry-run] ${label}を使用します: ${command}`);
    return;
  }
  if (!hasCommand(command)) {
    throw new Error(`${label}が見つかりません。PATHとインストール状態を確認してください: ${command}`);
  }
}

function composeArgs(extra = []) {
  return ['compose', '-f', resolve(DOCKER_DIR, 'compose.yml'), ...extra];
}

function runEnvSetup() {
  if (currentOptions.skipEnv) {
    console.log('ℹ️ --skip-env が指定されたため.env設定を省略します。');
    if (!currentOptions.dryRun && !existsSync(resolve(DOCKER_DIR, '.env'))) {
      throw new Error(`--skip-env が指定されていますが.envが見つかりません: ${resolve(DOCKER_DIR, '.env')}`);
    }
    return;
  }
  const envFile = resolve(DOCKER_DIR, '.env');
  const scriptArgs = currentOptions.forceEnv ? ['--force'] : [];
  if (isWindows) {
    requireCommand('powershell.exe', 'PowerShell');
    run('powershell.exe', ['-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', resolve(DOCKER_DIR, 'setup-env.ps1'), ...(currentOptions.forceEnv ? ['-Force'] : [])]);
  } else {
    requireCommand('bash', 'Bash');
    run('bash', [resolve(DOCKER_DIR, 'setup-env.sh'), ...scriptArgs]);
  }
  if (!currentOptions.dryRun && !existsSync(envFile)) {
    throw new Error(`.envが作成されませんでした: ${envFile}`);
  }
}

function checkDocker() {
  requireCommand('docker', 'Docker');
  run('docker', ['info']);
  run('docker', ['compose', 'version']);
}

function setupDocker() {
  checkDocker();
  runEnvSetup();
  const upArgs = ['up', '-d'];
  if (!currentOptions.noBuild) upArgs.push('--build');
  run('docker', composeArgs(upArgs));
  run('docker', composeArgs(['ps']));
  console.log('\n✅ Docker構築が完了しました。ログ確認:');
  console.log('  docker compose -f windows/docker/compose.yml logs -f');
}

function setupBuild() {
  requireCommand('bash', 'Bash');
  requireCommand('git', 'Git');
  requireCommand('go', 'Go');
  requireCommand('cmake', 'CMake');
  requireCommand('make', 'Make');
  requireCommand('patch', 'patch');
  run('bash', [resolve(ROOT, 'build', 'build.sh')], { env: { TARGET_OS: 'linux', TARGET_ARCH: 'amd64' } });
  console.log('\n✅ Linux amd64ビルドが完了しました: build/dist/yomiage-keiryou-amd64');
}

function setupVerify() {
  requireCommand('node', 'Node.js');
  requireCommand('git', 'Git');
  if (isWindows) {
    requireCommand('powershell.exe', 'PowerShell');
  } else {
    requireCommand('bash', 'Bash');
  }
  run('git', ['diff', '--check']);
  run('node', ['--check', resolve(ROOT, 'scripts', 'setup.mjs')]);
  run(isWindows ? 'powershell.exe' : 'bash', isWindows
    ? ['-NoProfile', '-ExecutionPolicy', 'Bypass', '-Command', "Get-ChildItem -Recurse -Filter *.ps1 | ForEach-Object { $tokens = $null; $errors = $null; [System.Management.Automation.Language.Parser]::ParseFile($_.FullName, [ref]$tokens, [ref]$errors) > $null; if ($errors.Count -gt 0) { $errors | ForEach-Object { Write-Error $_ }; exit 1 } }"]
    : ['-lc', "find . -type f -name '*.sh' -print0 | xargs -0 -r -n1 bash -n"],
  );
  console.log('\n✅ 自動化スクリプトと既存シェル構成の検証が完了しました。');
}

let currentOptions = { dryRun: false };

try {
  currentOptions = parseArgs(process.argv.slice(2));
  if (currentOptions.help) {
    usage();
    process.exit(0);
  }
  if (currentOptions.mode === 'docker') setupDocker();
  else if (currentOptions.mode === 'build') setupBuild();
  else setupVerify();
} catch (error) {
  console.error(`\n❌ ${error instanceof Error ? error.message : String(error)}`);
  console.error('詳細は --help で確認できます。');
  process.exit(1);
}
