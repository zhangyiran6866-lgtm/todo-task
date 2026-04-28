import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ErrorCode,
  ListToolsRequestSchema,
  McpError,
} from '@modelcontextprotocol/sdk/types.js';
import axios from 'axios';
import { exec } from 'child_process';
import { promisify } from 'util';
import path from 'path';
import fs from 'fs/promises';
import https from 'https';
import { fileURLToPath } from 'url';

const execAsync = promisify(exec);

const APIFOX_URL = 'https://api.apifox.com/v1';
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

type ImportFormat = 'openapi3' | 'swagger2';

function normalizePathForLog(inputPath: string): string {
  return inputPath.replace(process.cwd(), '.');
}

async function pathExists(targetPath: string): Promise<boolean> {
  try {
    await fs.access(targetPath);
    return true;
  } catch {
    return false;
  }
}

async function locateBackendPath(): Promise<string> {
  const candidates = [
    path.resolve(process.cwd(), 'packages/backend'),
    path.resolve(process.cwd(), '../../packages/backend'),
    path.resolve(__dirname, '../../packages/backend'),
    path.resolve(__dirname, '../../../packages/backend'),
  ];

  for (const candidate of candidates) {
    const mainPath = path.join(candidate, 'cmd/server/main.go');
    const swaggerPath = path.join(candidate, 'docs/swagger.json');
    if (await pathExists(mainPath) || await pathExists(swaggerPath)) {
      return candidate;
    }
  }

  throw new Error(
    `Cannot locate backend project path. Checked: ${candidates.map(normalizePathForLog).join(', ')}`
  );
}

function detectImportFormat(swaggerDoc: unknown): ImportFormat {
  if (swaggerDoc && typeof swaggerDoc === 'object') {
    const doc = swaggerDoc as Record<string, unknown>;
    if (typeof doc.openapi === 'string' && doc.openapi.startsWith('3')) {
      return 'openapi3';
    }
    if (typeof doc.swagger === 'string' && doc.swagger.startsWith('2')) {
      return 'swagger2';
    }
  }
  // Default to Swagger 2 because swaggo currently emits Swagger 2.0.
  return 'swagger2';
}

class ApifoxMcpServer {
  private server: Server;

  constructor() {
    this.server = new Server(
      { name: 'mcp-apifox-sync', version: '1.0.0' },
      { capabilities: { tools: {} } }
    );

    this.setupToolHandlers();
    
    // Error handling
    this.server.onerror = (error) => console.error('[MCP Error]', error);
    process.on('SIGINT', async () => {
      await this.server.close();
      process.exit(0);
    });
  }

  private setupToolHandlers() {
    this.server.setRequestHandler(ListToolsRequestSchema, async () => ({
      tools: [
        {
          name: 'sync_apifox',
          description: 'Generates OpenAPI spec from Go backend and uploads it to Apifox Luludina project',
          inputSchema: {
            type: 'object',
            properties: {
              apifoxToken: {
                type: 'string',
                description: 'Personal Access Token from Apifox API'
              },
              projectId: {
                type: 'string',
                description: 'The Project ID in Apifox (NOT the Team ID). Find it in Project Settings -> Basic Settings.'
              },
              options: {
                type: 'object',
                properties: {
                  tls: {
                    type: 'object',
                    properties: {
                      caCertPath: {
                        type: 'string',
                        description: 'Optional CA certificate path (PEM) for TLS verification in custom network environments.'
                      },
                      allowInsecureTLS: {
                        type: 'boolean',
                        description: 'Temporary local troubleshooting only. Disables TLS certificate verification.'
                      }
                    }
                  }
                }
              }
            },
            required: ['apifoxToken', 'projectId']
          }
        }
      ]
    }));

    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      if (request.params.name !== 'sync_apifox') {
        throw new McpError(ErrorCode.MethodNotFound, `Unknown tool: ${request.params.name}`);
      }

      const args = request.params.arguments as {
        apifoxToken?: string;
        projectId?: string;
        options?: {
          tls?: {
            caCertPath?: string;
            allowInsecureTLS?: boolean;
          };
        };
      };
      if (!args.apifoxToken || !args.projectId) {
        throw new McpError(ErrorCode.InvalidParams, 'apifoxToken and projectId arguments are required');
      }

      const token = args.apifoxToken;
      const projectId = args.projectId;
      const tlsOptions = args.options?.tls ?? {};

      try {
        const backendPath = await locateBackendPath();
        console.error(`Running swag init in ${normalizePathForLog(backendPath)}...`);
        
        try {
          const { stdout, stderr } = await execAsync('export PATH=$PATH:/usr/local/go/bin && go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go --parseDependency --parseInternal', { cwd: backendPath });
          console.error(stdout);
        } catch (execErr: any) {
          console.error('[swag init error]', execErr.stdout);
          console.error('[swag init error]', execErr.stderr);
          throw new Error('Failed to run swag init. Is swag installed?');
        }

        const swaggerPath = path.join(backendPath, 'docs/swagger.json');
        const swaggerData = await fs.readFile(swaggerPath, 'utf-8');
        const swaggerDoc = JSON.parse(swaggerData);
        const importFormat = detectImportFormat(swaggerDoc);

        let httpsAgent: https.Agent | undefined;
        if (tlsOptions.caCertPath) {
          const ca = await fs.readFile(tlsOptions.caCertPath, 'utf-8');
          httpsAgent = new https.Agent({
            rejectUnauthorized: !tlsOptions.allowInsecureTLS,
            ca,
          });
        } else if (tlsOptions.allowInsecureTLS) {
          httpsAgent = new https.Agent({ rejectUnauthorized: false });
        }

        const axiosClient = axios.create({
          headers: {
            'Authorization': `Bearer ${token}`,
            'X-Apifox-Api-Version': '2024-03-28'
          },
          httpsAgent,
        });

        console.error(`Uploading ${importFormat} data to project ${projectId}...`);
        
        // Exact Apifox payload required
        const importPayload = {
            format: importFormat,
            input: swaggerData,
            apiOverwriteBehavior: 'UPDATE_ALL'
        };

        const importResponse = await axiosClient.post(`${APIFOX_URL}/projects/${projectId}/import-openapi`, importPayload);

        return {
          content: [
            {
              type: 'text',
              text: `Successfully synced Swagger Docs to Apifox Project (ID: ${projectId}). Import result: ${JSON.stringify(importResponse.data?.data?.counters || importResponse.data)}`
            }
          ]
        };

      } catch (error: any) {
        const message = error?.message || '';
        if (message.includes('unable to get local issuer certificate') || message.includes('self signed certificate')) {
          return {
            content: [
              {
                type: 'text',
                text: [
                  'Failed to sync to Apifox: TLS certificate validation failed.',
                  'Fix options:',
                  '1) Provide options.tls.caCertPath with a valid root/intermediate CA PEM file.',
                  '2) For local troubleshooting only, set options.tls.allowInsecureTLS=true.',
                ].join(' '),
              },
            ],
            isError: true,
          };
        }

        console.error('[Apifox Sync Error]', error.response?.data || error.message);
        return {
          content: [
            {
              type: 'text',
              text: `Failed to sync to Apifox: ${error.response?.data ? JSON.stringify(error.response.data) : error.message}`
            }
          ],
          isError: true
        };
      }
    });
  }

  async run() {
    const transport = new StdioServerTransport();
    await this.server.connect(transport);
    console.error('Apifox MCP server running on stdio');
  }
}

const server = new ApifoxMcpServer();
server.run().catch(console.error);
