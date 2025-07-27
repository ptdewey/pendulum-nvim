use std::collections::HashMap;
use std::fs::{File, OpenOptions, create_dir_all};
use std::path::Path;
use std::process::Command;

use anyhow::{Result, anyhow};
use clap::Parser;
use csv::{ReaderBuilder, WriterBuilder};
use log::{error, info, warn};
use serde::{Deserialize, Serialize};
use serde_json::Value;
use tower_lsp::jsonrpc::{Error, Result as LspResult};
use tower_lsp::lsp_types::*;
use tower_lsp::{Client, LanguageServer, LspService, Server};

#[derive(Parser, Debug)]
#[command(name = "pendulum-lsp")]
#[command(about = "Pendulum time tracking LSP server")]
struct Args {
    /// Path to the CSV log file
    #[arg(long, short)]
    csv_path: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct ActivityData {
    active: String,
    branch: String,
    cwd: String,
    file: String,
    filetype: String,
    project: String,
    time: String,
}

struct PendulumLsp {
    client: Client,
    csv_file_path: String,
}

impl PendulumLsp {
    fn new(client: Client, csv_path: String) -> Self {
        Self {
            client,
            csv_file_path: csv_path,
        }
    }

    fn get_git_project(cwd: &str) -> String {
        match Command::new("git")
            .args(["config", "--local", "remote.origin.url"])
            .current_dir(cwd)
            .output()
        {
            Ok(output) => {
                let url = String::from_utf8_lossy(&output.stdout);
                let trimmed = url.trim();
                if let Some(captures) = regex::Regex::new(r".*/([^.]+)\.git$")
                    .unwrap()
                    .captures(trimmed)
                {
                    captures[1].to_string()
                } else {
                    "unknown_project".to_string()
                }
            }
            Err(_) => "unknown_project".to_string(),
        }
    }

    fn get_git_branch(cwd: &str) -> String {
        match Command::new("git")
            .args(["branch", "--show-current"])
            .current_dir(cwd)
            .output()
        {
            Ok(output) => {
                let branch = String::from_utf8_lossy(&output.stdout);
                let trimmed = branch.trim();
                if trimmed.is_empty() || trimmed.starts_with("fatal:") {
                    "unknown_branch".to_string()
                } else {
                    trimmed.to_string()
                }
            }
            Err(_) => "unknown_branch".to_string(),
        }
    }

    async fn write_csv_data(&self, data: &ActivityData) -> Result<()> {
        let path = &self.csv_file_path;

        // Create directory if it doesn't exist
        if let Some(parent) = Path::new(path).parent() {
            create_dir_all(parent)?;
        }

        let file_exists = Path::new(path).exists();
        let mut file_empty = false;

        if file_exists {
            let metadata = std::fs::metadata(path)?;
            file_empty = metadata.len() == 0;
        }

        let file = OpenOptions::new().create(true).append(true).open(path)?;

        let mut writer = WriterBuilder::new().from_writer(file);

        // Write header if file is new or empty
        if !file_exists || file_empty {
            writer.write_record([
                "active", "branch", "cwd", "file", "filetype", "project", "time",
            ])?;
        }

        // Write data row
        writer.write_record([
            &data.active,
            &data.branch,
            &data.cwd,
            &data.file,
            &data.filetype,
            &data.project,
            &data.time,
        ])?;

        writer.flush()?;
        Ok(())
    }
}

#[tower_lsp::async_trait]
impl LanguageServer for PendulumLsp {
    async fn initialize(&self, _params: InitializeParams) -> LspResult<InitializeResult> {
        info!(
            "Pendulum LSP initializing with CSV path: {}",
            self.csv_file_path
        );

        Ok(InitializeResult {
            capabilities: ServerCapabilities {
                execute_command_provider: Some(ExecuteCommandOptions {
                    commands: vec!["pendulum.logActivity".to_string()],
                    ..Default::default()
                }),
                ..Default::default()
            },
            server_info: Some(ServerInfo {
                name: "Pendulum LSP".to_string(),
                version: Some("0.1.0".to_string()),
            }),
        })
    }

    async fn initialized(&self, _: InitializedParams) {
        info!("Pendulum LSP initialized");
        self.client
            .log_message(MessageType::INFO, "Pendulum LSP server initialized")
            .await;
    }

    async fn shutdown(&self) -> LspResult<()> {
        info!("Pendulum LSP shutting down");
        Ok(())
    }

    async fn execute_command(&self, params: ExecuteCommandParams) -> LspResult<Option<Value>> {
        match params.command.as_str() {
            "pendulum.logActivity" => {
                if let Some(args) = params.arguments.first() {
                    match serde_json::from_value::<HashMap<String, Value>>(args.clone()) {
                        Ok(data) => {
                            let activity = ActivityData {
                                time: data
                                    .get("time")
                                    .and_then(|v| v.as_str())
                                    .unwrap_or("")
                                    .to_string(),
                                active: data
                                    .get("active")
                                    .and_then(|v| v.as_str())
                                    .unwrap_or("false")
                                    .to_string(),
                                file: data
                                    .get("file")
                                    .and_then(|v| v.as_str())
                                    .unwrap_or("")
                                    .to_string(),
                                filetype: data
                                    .get("filetype")
                                    .and_then(|v| v.as_str())
                                    .unwrap_or("unknown_filetype")
                                    .to_string(),
                                cwd: data
                                    .get("cwd")
                                    .and_then(|v| v.as_str())
                                    .unwrap_or("")
                                    .to_string(),
                                project: Self::get_git_project(
                                    data.get("cwd").and_then(|v| v.as_str()).unwrap_or(""),
                                ),
                                branch: Self::get_git_branch(
                                    data.get("cwd").and_then(|v| v.as_str()).unwrap_or(""),
                                ),
                            };

                            match self.write_csv_data(&activity).await {
                                Ok(_) => {
                                    info!("Activity logged successfully");
                                    Ok(Some(Value::Bool(true)))
                                }
                                Err(e) => {
                                    error!("Failed to write CSV data: {e}");
                                    self.client
                                        .log_message(
                                            MessageType::ERROR,
                                            format!("Failed to write CSV data: {e}"),
                                        )
                                        .await;
                                    Err(Error::internal_error())
                                }
                            }
                        }
                        Err(e) => {
                            error!("Failed to parse activity data: {e}");
                            Err(Error::invalid_params("Invalid activity data format"))
                        }
                    }
                } else {
                    Err(Error::invalid_params("Missing activity data"))
                }
            }
            _ => Err(Error::method_not_found()),
        }
    }
}

#[tokio::main]
async fn main() {
    env_logger::init();

    let args = Args::parse();

    let stdin = tokio::io::stdin();
    let stdout = tokio::io::stdout();

    let (service, socket) =
        LspService::build(|client| PendulumLsp::new(client, args.csv_path)).finish();
    Server::new(stdin, stdout, socket).serve(service).await;
}
