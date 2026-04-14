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

const execAsync = promisify(exec);

const APIFOX_URL = 'https://api.apifox.com/v1';

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

      const args = request.params.arguments as { apifoxToken?: string, projectId?: string };
      if (!args.apifoxToken || !args.projectId) {
        throw new McpError(ErrorCode.InvalidParams, 'apifoxToken and projectId arguments are required');
      }

      const token = args.apifoxToken;
      const projectId = args.projectId;

      try {
        const backendPath = path.resolve(process.cwd(), '../../packages/backend'); 
        console.error(`Running swag init in ${backendPath}...`);
        
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

        const axiosClient = axios.create({
          headers: {
            'Authorization': `Bearer ${token}`,
            'X-Apifox-Api-Version': '2024-03-28'
          }
        });

        console.error(`Uploading swagger data to project ${projectId}...`);
        
        // Exact Apifox payload required
        const importPayload = {
            format: 'openapi3',
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
