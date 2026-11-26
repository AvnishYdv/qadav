const vscode = require('vscode');
const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

/**
 * Qadav Language Support Extension
 * Zero-npm LSP client launching bin/qyn.exe lsp over stdio
 */
let lspProcess = null;
let diagnosticCollection = null;
let outputChannel = null;

function activate(context) {
    outputChannel = vscode.window.createOutputChannel('Qadav');
    outputChannel.appendLine('[Qadav] Extension activated.');
    diagnosticCollection = vscode.languages.createDiagnosticCollection('qadav');
    context.subscriptions.push(diagnosticCollection);

    // Register commands
    context.subscriptions.push(
        vscode.commands.registerCommand('qadav.build', () => runCompilerCommand('build')),
        vscode.commands.registerCommand('qadav.run', () => runCompilerCommand('run')),
        vscode.commands.registerCommand('qadav.typecheck', () => runCompilerCommand('check')),
        vscode.commands.registerCommand('qadav.doctor', () => runCompilerCommand('doctor'))
    );

    // Start LSP if enabled
    const config = vscode.workspace.getConfiguration('qadav');
    if (config.get('enableLsp', true)) {
        startLspServer(context);
    }
}

function getCompilerPath() {
    const config = vscode.workspace.getConfiguration('qadav');
    const configuredPath = config.get('compilerPath', 'bin/qyn.exe');
    if (path.isAbsolute(configuredPath)) {
        return configuredPath;
    }
    const workspaceFolders = vscode.workspace.workspaceFolders;
    if (workspaceFolders && workspaceFolders.length > 0) {
        const rootPath = workspaceFolders[0].uri.fsPath;
        const resolved = path.join(rootPath, configuredPath);
        if (fs.existsSync(resolved)) {
            return resolved;
        }
    }
    return configuredPath;
}

function startLspServer(context) {
    const compiler = getCompilerPath();
    outputChannel.appendLine(`[Qadav LSP] Starting Language Server via: ${compiler} lsp`);
    try {
        lspProcess = spawn(compiler, ['lsp'], {
            stdio: ['pipe', 'pipe', 'pipe']
        });

        lspProcess.stdout.on('data', (data) => {
            outputChannel.appendLine(`[Qadav LSP RX] ${data.toString()}`);
        });

        lspProcess.stderr.on('data', (data) => {
            outputChannel.appendLine(`[Qadav LSP ERR] ${data.toString()}`);
        });

        lspProcess.on('exit', (code) => {
            outputChannel.appendLine(`[Qadav LSP] Server exited with code: ${code}`);
        });
    } catch (err) {
        outputChannel.appendLine(`[Qadav LSP] Failed to launch server: ${err.message}`);
    }
}

function runCompilerCommand(action) {
    const editor = vscode.window.activeTextEditor;
    const compiler = getCompilerPath();
    const docPath = editor ? editor.document.uri.fsPath : '';

    if (action === 'doctor') {
        const proc = spawn(compiler, ['doctor']);
        let out = '';
        proc.stdout.on('data', (d) => { out += d.toString(); });
        proc.on('close', () => {
            outputChannel.appendLine(out);
            outputChannel.show(true);
        });
        return;
    }

    if (!docPath) {
        vscode.window.showErrorMessage('No active Qadav (.ydv) file open in editor.');
        return;
    }

    if (action === 'check') {
        const proc = spawn(compiler, ['check', docPath]);
        proc.stdout.on('data', (d) => outputChannel.appendLine(d.toString()));
        proc.on('close', (code) => {
            if (code === 0) {
                vscode.window.showInformationMessage('Qadav: Typecheck passed with 0 errors.');
            } else {
                vscode.window.showErrorMessage('Qadav: Typecheck failed.');
            }
        });
    } else if (action === 'build') {
        const outExe = docPath.replace(/\.ydv$/, '.exe');
        const proc = spawn(compiler, ['build', docPath, '-o', outExe]);
        proc.stdout.on('data', (d) => outputChannel.appendLine(d.toString()));
        proc.on('close', (code) => {
            if (code === 0) {
                vscode.window.showInformationMessage(`Qadav: Built standalone executable ${path.basename(outExe)}`);
            } else {
                vscode.window.showErrorMessage('Qadav: Build failed.');
            }
        });
    } else if (action === 'run') {
        const outExe = docPath.replace(/\.ydv$/, '.exe');
        const buildProc = spawn(compiler, ['build', docPath, '-o', outExe]);
        buildProc.on('close', (code) => {
            if (code === 0) {
                const runProc = spawn(outExe, []);
                runProc.stdout.on('data', (d) => outputChannel.appendLine(d.toString()));
                runProc.on('close', () => outputChannel.show(true));
            } else {
                vscode.window.showErrorMessage('Qadav: Build before run failed.');
            }
        });
    }
}

function deactivate() {
    if (lspProcess) {
        lspProcess.kill();
        lspProcess = null;
    }
}

module.exports = {
    activate,
    deactivate
};