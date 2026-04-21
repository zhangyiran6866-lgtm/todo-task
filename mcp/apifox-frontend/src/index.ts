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
import fs from 'fs/promises';
import path from 'path';
import { promisify } from 'util';

const execAsync = promisify(exec);
const APIFOX_URL = 'https://api.apifox.com/v1';
const GENERATED_START = (moduleName: string) => `// <apifox-generated module="${moduleName}">`;
const GENERATED_END = (moduleName: string) => `// </apifox-generated module="${moduleName}">`;

type HttpMethod = 'get' | 'post' | 'put' | 'patch' | 'delete';
type WriteMode = 'dry-run' | 'write';
type ScopeType = 'module' | 'path-prefix' | 'operation-ids';
type ApiStatus = 'added' | 'updated' | 'unchanged' | 'removed';
type FileStatus = 'created' | 'updated' | 'unchanged' | 'skipped';

interface SyncArguments {
  apifoxToken?: string;
  projectId?: string;
  scope?: {
    type?: ScopeType;
    value?: string | string[];
  };
  frontend?: {
    apiDir?: string;
    targetFile?: string;
    requestImportPath?: string;
    basePath?: string;
  };
  writeMode?: WriteMode;
  updateStrategy?: 'generated-block';
  options?: {
    includeTypes?: boolean;
    includeRequestFunctions?: boolean;
    includeIndexExport?: boolean;
    allowInsertGeneratedBlock?: boolean;
    runTypeCheck?: boolean;
    buildCommand?: string;
    quoteStyle?: 'single' | 'double';
    semicolons?: boolean;
  };
}

interface ResolvedSyncArguments {
  apifoxToken: string;
  projectId: string;
  scope: {
    type: ScopeType;
    value: string | string[];
  };
  frontend: {
    apiDir: string;
    targetFile: string;
    requestImportPath: string;
    basePath: string;
  };
  writeMode: WriteMode;
  updateStrategy: 'generated-block';
  options: {
    includeTypes: boolean;
    includeRequestFunctions: boolean;
    includeIndexExport: boolean;
    allowInsertGeneratedBlock: boolean;
    runTypeCheck: boolean;
    buildCommand: string;
    quoteStyle: 'single' | 'double';
    semicolons: boolean;
  };
}

interface ModuleConfig {
  moduleName: string;
  pathPrefixes: string[];
  targetFile: string;
  apiObjectName: string;
}

interface OpenApiDocument {
  openapi?: string;
  swagger?: string;
  paths?: Record<string, PathItem>;
  components?: {
    schemas?: Record<string, OpenApiSchema>;
  };
  definitions?: Record<string, OpenApiSchema>;
}

type PathItem = Partial<Record<HttpMethod, OpenApiOperation>>;

interface OpenApiOperation {
  operationId?: string;
  summary?: string;
  description?: string;
  tags?: string[];
  parameters?: OpenApiParameter[];
  requestBody?: {
    required?: boolean;
    content?: Record<string, { schema?: OpenApiSchema }>;
  };
  responses?: Record<string, OpenApiResponse>;
  security?: Array<Record<string, string[]>>;
}

interface OpenApiParameter {
  name: string;
  in: 'path' | 'query' | 'header' | 'cookie';
  required?: boolean;
  schema?: OpenApiSchema;
  description?: string;
}

interface OpenApiResponse {
  description?: string;
  content?: Record<string, { schema?: OpenApiSchema }>;
}

interface OpenApiSchema {
  $ref?: string;
  type?: string | string[];
  format?: string;
  enum?: Array<string | number | boolean | null>;
  nullable?: boolean;
  required?: string[];
  properties?: Record<string, OpenApiSchema>;
  items?: OpenApiSchema;
  allOf?: OpenApiSchema[];
  oneOf?: OpenApiSchema[];
  anyOf?: OpenApiSchema[];
  additionalProperties?: boolean | OpenApiSchema;
  description?: string;
}

interface NormalizedOperation {
  operationId: string;
  moduleName: string;
  method: HttpMethod;
  rawPath: string;
  frontendPath: string;
  summary?: string;
  tags: string[];
  requiresAuth: boolean;
  pathParams: OpenApiParameter[];
  queryParams: OpenApiParameter[];
  requestBody?: OpenApiSchema;
  responseData?: OpenApiSchema;
  requestTypeName?: string;
  responseTypeName: string;
  functionName: string;
}

interface WarningItem {
  code: string;
  message: string;
  api?: string;
}

interface ApiReportItem {
  method: string;
  path: string;
  functionName: string;
  requestType?: string;
  responseType?: string;
  status: ApiStatus;
}

interface TypeReportItem {
  name: string;
  status: ApiStatus;
}

interface FileReportItem {
  path: string;
  status: FileStatus;
}

interface BuildReport {
  executed: boolean;
  success?: boolean;
  command?: string;
  summary?: string;
}

interface SyncReport {
  success: boolean;
  mode: WriteMode;
  scope: {
    type: ScopeType;
    value: string | string[];
    matchedApiCount: number;
  };
  target: {
    file: string;
    moduleName: string;
  };
  apis: ApiReportItem[];
  types: TypeReportItem[];
  files: FileReportItem[];
  diff?: string;
  warnings: WarningItem[];
  typeCheck?: BuildReport;
  nextSteps: string[];
}

interface GeneratedCode {
  body: string;
  apiItems: ApiReportItem[];
  typeItems: TypeReportItem[];
}

const MODULES: Record<string, ModuleConfig> = {
  auth: {
    moduleName: 'auth',
    pathPrefixes: ['/auth'],
    targetFile: 'packages/frontend/src/api/auth.ts',
    apiObjectName: 'authApi',
  },
  user: {
    moduleName: 'user',
    pathPrefixes: ['/users'],
    targetFile: 'packages/frontend/src/api/user.ts',
    apiObjectName: 'userApi',
  },
  task: {
    moduleName: 'task',
    pathPrefixes: ['/tasks'],
    targetFile: 'packages/frontend/src/api/task.ts',
    apiObjectName: 'taskApi',
  },
};

class ApifoxFrontendMcpServer {
  private server: Server;

  constructor() {
    this.server = new Server(
      { name: 'mcp-apifox-frontend', version: '1.0.0' },
      { capabilities: { tools: {} } },
    );

    this.setupToolHandlers();
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
          name: 'sync_frontend_api_from_apifox',
          description:
            'Synchronize a clearly scoped Apifox API contract into frontend TypeScript API files. Does not modify Vue views, stores, routes, or business logic.',
          inputSchema: {
            type: 'object',
            properties: {
              apifoxToken: {
                type: 'string',
                description: 'Apifox Personal Access Token. It is used only for this call and is never persisted.',
              },
              projectId: {
                type: 'string',
                description: 'Apifox Project ID, not Team ID.',
              },
              scope: {
                type: 'object',
                description: 'Explicit API sync scope.',
                properties: {
                  type: {
                    type: 'string',
                    enum: ['module', 'path-prefix', 'operation-ids'],
                  },
                  value: {
                    oneOf: [
                      { type: 'string' },
                      { type: 'array', items: { type: 'string' } },
                    ],
                  },
                },
                required: ['type', 'value'],
              },
              frontend: {
                type: 'object',
                properties: {
                  apiDir: { type: 'string', description: 'Default: packages/frontend/src/api' },
                  targetFile: { type: 'string', description: 'Optional explicit target file.' },
                  requestImportPath: { type: 'string', description: 'Default: @/utils/request' },
                  basePath: { type: 'string', description: 'Default: /api/v1' },
                },
              },
              writeMode: {
                type: 'string',
                enum: ['dry-run', 'write'],
                description: 'Default: dry-run.',
              },
              updateStrategy: {
                type: 'string',
                enum: ['generated-block'],
                description: 'Only generated-block is supported in this version.',
              },
              options: {
                type: 'object',
                properties: {
                  includeTypes: { type: 'boolean' },
                  includeRequestFunctions: { type: 'boolean' },
                  includeIndexExport: { type: 'boolean' },
                  allowInsertGeneratedBlock: { type: 'boolean' },
                  runTypeCheck: { type: 'boolean' },
                  buildCommand: { type: 'string' },
                  quoteStyle: { type: 'string', enum: ['single', 'double'] },
                  semicolons: { type: 'boolean' },
                },
              },
            },
            required: ['apifoxToken', 'projectId', 'scope'],
          },
        },
      ],
    }));

    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      if (request.params.name !== 'sync_frontend_api_from_apifox') {
        throw new McpError(ErrorCode.MethodNotFound, `Unknown tool: ${request.params.name}`);
      }

      try {
        const args = validateArguments(request.params.arguments);
        const repoRoot = await findRepoRoot(process.cwd());
        const report = await syncFrontendApi(args, repoRoot);

        return {
          content: [
            {
              type: 'text',
              text: JSON.stringify(report, null, 2),
            },
          ],
          isError: !report.success,
        };
      } catch (error: unknown) {
        const message = error instanceof Error ? error.message : 'Unknown sync error';
        return {
          content: [
            {
              type: 'text',
              text: JSON.stringify(
                {
                  success: false,
                  message,
                },
                null,
                2,
              ),
            },
          ],
          isError: true,
        };
      }
    });
  }

  async run() {
    const transport = new StdioServerTransport();
    await this.server.connect(transport);
    console.error('Apifox Frontend MCP server running on stdio');
  }
}

function validateArguments(raw: unknown): ResolvedSyncArguments {
  if (!isObject(raw)) {
    throw new McpError(ErrorCode.InvalidParams, 'Arguments must be an object');
  }

  const args = raw as SyncArguments;
  if (!args.apifoxToken || !args.projectId) {
    throw new McpError(ErrorCode.InvalidParams, 'apifoxToken and projectId are required');
  }
  if (!args.scope?.type || args.scope.value === undefined) {
    throw new McpError(
      ErrorCode.InvalidParams,
      'scope is required. Use module, path-prefix, or operation-ids.',
    );
  }
  if (!['module', 'path-prefix', 'operation-ids'].includes(args.scope.type)) {
    throw new McpError(ErrorCode.InvalidParams, 'scope.type must be module, path-prefix, or operation-ids');
  }
  if (Array.isArray(args.scope.value) && args.scope.value.length === 0) {
    throw new McpError(ErrorCode.InvalidParams, 'scope.value must not be an empty array');
  }
  if (args.updateStrategy && args.updateStrategy !== 'generated-block') {
    throw new McpError(ErrorCode.InvalidParams, 'Only updateStrategy=generated-block is supported');
  }

  return {
    apifoxToken: args.apifoxToken,
    projectId: args.projectId,
    scope: {
      type: args.scope.type,
      value: args.scope.value,
    },
    frontend: {
      apiDir: args.frontend?.apiDir ?? 'packages/frontend/src/api',
      targetFile: args.frontend?.targetFile ?? '',
      requestImportPath: args.frontend?.requestImportPath ?? '@/utils/request',
      basePath: args.frontend?.basePath ?? '/api/v1',
    },
    writeMode: args.writeMode ?? 'dry-run',
    updateStrategy: 'generated-block',
    options: {
      includeTypes: args.options?.includeTypes ?? true,
      includeRequestFunctions: args.options?.includeRequestFunctions ?? true,
      includeIndexExport: args.options?.includeIndexExport ?? false,
      allowInsertGeneratedBlock: args.options?.allowInsertGeneratedBlock ?? false,
      runTypeCheck: args.options?.runTypeCheck ?? false,
      buildCommand: args.options?.buildCommand ?? 'pnpm --filter frontend build',
      quoteStyle: args.options?.quoteStyle ?? 'single',
      semicolons: args.options?.semicolons ?? false,
    },
  };
}

async function syncFrontendApi(args: ResolvedSyncArguments, repoRoot: string): Promise<SyncReport> {
  const warnings: WarningItem[] = [];
  const openApi = await fetchOpenApi(args.apifoxToken, args.projectId);
  const moduleConfig = resolveModuleConfig(args);
  const operations = normalizeOperations(openApi, args.frontend.basePath, moduleConfig, warnings);
  const matched = filterOperationsByScope(operations, args.scope);

  if (matched.length === 0) {
    return failureReport(args, moduleConfig, 'NO_MATCHED_APIS', 'No APIs matched the requested scope');
  }

  const targetFile = resolveTargetFile(args, moduleConfig);
  ensureSafeTarget(repoRoot, targetFile);

  const generated = generateCode(matched, moduleConfig, args, openApi, warnings);
  const writeResult = await applyGeneratedBlock({
    repoRoot,
    targetFile,
    moduleName: moduleConfig.moduleName,
    body: generated.body,
    mode: args.writeMode,
    allowInsertGeneratedBlock: args.options.allowInsertGeneratedBlock,
  });

  const buildReport =
    args.writeMode === 'write' && args.options.runTypeCheck
      ? await runTypeCheck(repoRoot, args.options.buildCommand)
      : { executed: false };

  const nextSteps = buildNextSteps(args, writeResult.status, buildReport);

  return {
    success: writeResult.status !== 'skipped' || args.writeMode === 'dry-run',
    mode: args.writeMode,
    scope: {
      type: args.scope.type,
      value: args.scope.value,
      matchedApiCount: matched.length,
    },
    target: {
      file: targetFile,
      moduleName: moduleConfig.moduleName,
    },
    apis: generated.apiItems,
    types: generated.typeItems,
    files: [
      {
        path: targetFile,
        status: writeResult.status,
      },
    ],
    diff: writeResult.diff,
    warnings: [...warnings, ...writeResult.warnings],
    typeCheck: buildReport,
    nextSteps,
  };
}

async function fetchOpenApi(apifoxToken: string, projectId: string): Promise<OpenApiDocument> {
  const client = axios.create({
    headers: {
      Authorization: `Bearer ${apifoxToken}`,
      'X-Apifox-Api-Version': '2024-03-28',
      'Content-Type': 'application/json',
    },
  });

  const response = await client.post<unknown>(
    `${APIFOX_URL}/projects/${encodeURIComponent(projectId)}/export-openapi?locale=zh-CN`,
    {
      scope: { type: 'ALL' },
      options: {
        includeApifoxExtensionProperties: false,
        addFoldersToTags: false,
      },
      oasVersion: '3.1',
      exportFormat: 'JSON',
    },
  );

  if (!isObject(response.data)) {
    throw new Error('Apifox OpenAPI export returned invalid data');
  }

  const document = response.data as OpenApiDocument;
  if (!document.paths || Object.keys(document.paths).length === 0) {
    throw new Error('Apifox OpenAPI export has no paths');
  }
  return document;
}

function resolveModuleConfig(args: ResolvedSyncArguments): ModuleConfig {
  if (args.scope.type === 'module' && typeof args.scope.value === 'string') {
    const config = MODULES[args.scope.value];
    if (!config) {
      throw new McpError(ErrorCode.InvalidParams, `Unknown module scope: ${args.scope.value}`);
    }
    return config;
  }

  const firstValue = Array.isArray(args.scope.value) ? args.scope.value[0] : args.scope.value;
  const normalized = normalizeFrontendPath(firstValue, args.frontend.basePath);
  const matched = Object.values(MODULES).find((item) =>
    item.pathPrefixes.some((prefix) => normalized.startsWith(prefix)),
  );

  if (matched) {
    return matched;
  }

  if (!args.frontend.targetFile) {
    throw new McpError(
      ErrorCode.InvalidParams,
      'frontend.targetFile is required when scope cannot be mapped to a built-in module',
    );
  }

  const moduleName = path.basename(args.frontend.targetFile, path.extname(args.frontend.targetFile));
  return {
    moduleName,
    pathPrefixes: [normalized],
    targetFile: args.frontend.targetFile,
    apiObjectName: `${camelCase(moduleName)}Api`,
  };
}

function normalizeOperations(
  openApi: OpenApiDocument,
  basePath: string,
  moduleConfig: ModuleConfig,
  warnings: WarningItem[],
): NormalizedOperation[] {
  const operations: NormalizedOperation[] = [];
  const paths = openApi.paths ?? {};

  for (const [rawPath, pathItem] of Object.entries(paths)) {
    for (const method of ['get', 'post', 'put', 'patch', 'delete'] as HttpMethod[]) {
      const operation = pathItem[method];
      if (!operation) continue;

      const frontendPath = normalizeFrontendPath(rawPath, basePath);
      const moduleName = resolveModuleName(frontendPath, moduleConfig);
      const functionName = resolveFunctionName(operation.operationId, method, frontendPath);
      const parameters = operation.parameters ?? [];
      const requestBody = getJsonSchema(operation.requestBody?.content);
      const responseData = unwrapResponseData(operation.responses, `${method.toUpperCase()} ${frontendPath}`, warnings);
      const responseTypeName = responseData ? resolveResponseTypeName(functionName, responseData) : 'void';
      const requestTypeName = requestBody ? `${pascalCase(functionName)}Req` : undefined;

      operations.push({
        operationId: operation.operationId ?? functionName,
        moduleName,
        method,
        rawPath,
        frontendPath,
        summary: operation.summary,
        tags: operation.tags ?? [],
        requiresAuth: operation.security !== undefined ? operation.security.length > 0 : !frontendPath.startsWith('/auth/login') && !frontendPath.startsWith('/auth/register'),
        pathParams: parameters.filter((param) => param.in === 'path'),
        queryParams: parameters.filter((param) => param.in === 'query'),
        requestBody,
        responseData,
        requestTypeName,
        responseTypeName,
        functionName,
      });
    }
  }

  return operations.sort((a, b) => `${a.frontendPath}:${a.method}`.localeCompare(`${b.frontendPath}:${b.method}`));
}

function filterOperationsByScope(operations: NormalizedOperation[], scope: ResolvedSyncArguments['scope']) {
  if (scope.type === 'module' && typeof scope.value === 'string') {
    const config = MODULES[scope.value];
    return operations.filter((operation) =>
      config?.pathPrefixes.some((prefix) => operation.frontendPath.startsWith(prefix)),
    );
  }

  if (scope.type === 'path-prefix') {
    const prefixes = asArray(scope.value).map((item) => normalizeFrontendPath(item, ''));
    return operations.filter((operation) => prefixes.some((prefix) => operation.frontendPath.startsWith(prefix)));
  }

  const operationIds = new Set(asArray(scope.value));
  return operations.filter((operation) => operationIds.has(operation.operationId));
}

function generateCode(
  operations: NormalizedOperation[],
  moduleConfig: ModuleConfig,
  args: ResolvedSyncArguments,
  openApi: OpenApiDocument,
  warnings: WarningItem[],
): GeneratedCode {
  const quote = args.options.quoteStyle === 'double' ? '"' : "'";
  const semi = args.options.semicolons ? ';' : '';
  const typeBlocks = new Map<string, string>();
  const typeItems: TypeReportItem[] = [];
  const apiItems: ApiReportItem[] = [];

  if (args.options.includeTypes) {
    for (const operation of operations) {
      if (operation.queryParams.length > 0) {
        const name = `${pascalCase(operation.functionName)}Req`;
        typeBlocks.set(name, renderParamsInterface(name, operation.queryParams, openApi, warnings, semi));
      }
      if (operation.requestBody && operation.requestTypeName) {
        typeBlocks.set(
          operation.requestTypeName,
          renderSchemaDeclaration(operation.requestTypeName, operation.requestBody, openApi, warnings, semi),
        );
      }
      if (operation.responseData && operation.responseTypeName !== 'void') {
        typeBlocks.set(
          operation.responseTypeName,
          renderSchemaDeclaration(operation.responseTypeName, operation.responseData, openApi, warnings, semi),
        );
      }
    }
  }

  for (const name of typeBlocks.keys()) {
    typeItems.push({ name, status: 'updated' });
  }

  const lines: string[] = [
    `import request from ${quote}${args.frontend.requestImportPath}${quote}${semi}`,
    '',
  ];

  if (args.options.includeTypes) {
    lines.push(...Array.from(typeBlocks.values()).flatMap((block) => [block, '']));
  }

  if (args.options.includeRequestFunctions) {
    lines.push(`export const ${moduleConfig.apiObjectName} = {`);
    operations.forEach((operation, index) => {
      lines.push(...renderApiFunction(operation, openApi, warnings, quote, semi));
      if (index < operations.length - 1) {
        lines.push('');
      }
    });
    lines.push(`}${semi}`);
  }

  for (const operation of operations) {
    apiItems.push({
      method: operation.method.toUpperCase(),
      path: operation.frontendPath,
      functionName: operation.functionName,
      requestType: operation.requestTypeName ?? (operation.queryParams.length > 0 ? `${pascalCase(operation.functionName)}Req` : undefined),
      responseType: operation.responseTypeName,
      status: 'updated',
    });
  }

  return {
    body: lines.join('\n').trimEnd() + '\n',
    apiItems,
    typeItems,
  };
}

function renderSchemaDeclaration(
  name: string,
  schema: OpenApiSchema,
  openApi: OpenApiDocument,
  warnings: WarningItem[],
  semi: string,
): string {
  const resolved = resolveRef(schema, openApi);
  if (resolved.enum) {
    return `export type ${name} = ${renderEnum(resolved.enum)}${semi}`;
  }
  if (isObjectSchema(resolved)) {
    return renderInterface(name, resolved, openApi, warnings, semi);
  }
  return `export type ${name} = ${schemaToTs(resolved, openApi, warnings)}${semi}`;
}

function renderParamsInterface(
  name: string,
  params: OpenApiParameter[],
  openApi: OpenApiDocument,
  warnings: WarningItem[],
  semi: string,
): string {
  const lines = [`export interface ${name} {`];
  for (const param of params) {
    const optional = param.required ? '' : '?';
    const type = param.schema ? schemaToTs(param.schema, openApi, warnings) : 'unknown';
    if (!param.schema) {
      warnings.push({
        code: 'PARAM_SCHEMA_MISSING',
        message: `${param.name} is missing schema, generated as unknown`,
      });
    }
    lines.push(`  ${safePropertyName(param.name)}${optional}: ${type}${semi}`);
  }
  lines.push('}');
  return lines.join('\n');
}

function renderInterface(
  name: string,
  schema: OpenApiSchema,
  openApi: OpenApiDocument,
  warnings: WarningItem[],
  semi: string,
): string {
  const resolved = resolveRef(schema, openApi);
  const required = new Set(resolved.required ?? []);
  const lines = [`export interface ${name} {`];
  const properties = resolved.properties ?? {};
  for (const [propertyName, propertySchema] of Object.entries(properties)) {
    const optional = required.has(propertyName) ? '' : '?';
    lines.push(`  ${safePropertyName(propertyName)}${optional}: ${schemaToTs(propertySchema, openApi, warnings)}${semi}`);
  }
  if (Object.keys(properties).length === 0) {
    lines.push(`  [key: string]: unknown${semi}`);
  }
  lines.push('}');
  return lines.join('\n');
}

function renderApiFunction(
  operation: NormalizedOperation,
  openApi: OpenApiDocument,
  warnings: WarningItem[],
  quote: string,
  semi: string,
): string[] {
  const params: string[] = [];
  for (const pathParam of operation.pathParams) {
    params.push(`${camelCase(pathParam.name)}: ${pathParam.schema ? schemaToTs(pathParam.schema, openApi, warnings) : 'string'}`);
  }
  if (operation.queryParams.length > 0) {
    params.push(`params?: ${pascalCase(operation.functionName)}Req`);
  }
  if (operation.requestBody && operation.requestTypeName) {
    params.push(`data: ${operation.requestTypeName}`);
  }

  const returnType = operation.responseTypeName === 'void' ? 'void' : operation.responseTypeName;
  const requestPath = renderRequestPath(operation.frontendPath, operation.pathParams, quote);
  const axiosArgs = buildAxiosArgs(operation, requestPath);

  return [
    `  ${operation.functionName}(${params.join(', ')}): Promise<${returnType}> {`,
    `    return request.${operation.method}(${axiosArgs}) as unknown as Promise<${returnType}>${semi}`,
    '  },',
  ];
}

function buildAxiosArgs(operation: NormalizedOperation, requestPath: string): string {
  if (operation.method === 'get') {
    return operation.queryParams.length > 0 ? `${requestPath}, { params }` : requestPath;
  }
  if (operation.method === 'delete') {
    return operation.queryParams.length > 0 ? `${requestPath}, { params }` : requestPath;
  }
  if (operation.requestBody) {
    return operation.queryParams.length > 0 ? `${requestPath}, data, { params }` : `${requestPath}, data`;
  }
  return operation.queryParams.length > 0 ? `${requestPath}, undefined, { params }` : requestPath;
}

function renderRequestPath(frontendPath: string, pathParams: OpenApiParameter[], quote: string): string {
  if (pathParams.length === 0) {
    return `${quote}${frontendPath}${quote}`;
  }

  let template = frontendPath;
  for (const param of pathParams) {
    template = template.replace(`{${param.name}}`, `\${${camelCase(param.name)}}`);
  }
  return `\`${template}\``;
}

async function applyGeneratedBlock(input: {
  repoRoot: string;
  targetFile: string;
  moduleName: string;
  body: string;
  mode: WriteMode;
  allowInsertGeneratedBlock: boolean;
}): Promise<{ status: FileStatus; diff?: string; warnings: WarningItem[] }> {
  const absoluteTarget = path.resolve(input.repoRoot, input.targetFile);
  const warnings: WarningItem[] = [];
  let existing = '';
  let exists = true;

  try {
    existing = await fs.readFile(absoluteTarget, 'utf-8');
  } catch (error: unknown) {
    if (isNodeError(error) && error.code === 'ENOENT') {
      exists = false;
    } else {
      throw error;
    }
  }

  const start = GENERATED_START(input.moduleName);
  const end = GENERATED_END(input.moduleName);
  const generatedBlock = `${start}\n${input.body}${end}\n`;
  const oldGenerated = extractGeneratedBlock(existing, start, end);
  let nextContent: string;

  if (!exists) {
    nextContent = generatedBlock;
  } else if (oldGenerated) {
    nextContent = existing.replace(oldGenerated, generatedBlock);
  } else if (input.allowInsertGeneratedBlock) {
    const separator = existing.endsWith('\n') ? '\n' : '\n\n';
    nextContent = `${existing}${separator}${generatedBlock}`;
  } else {
    warnings.push({
      code: 'GENERATED_BLOCK_MISSING',
      message:
        'Target file exists without an apifox-generated block. Skipped write unless options.allowInsertGeneratedBlock=true.',
    });
    return {
      status: input.mode === 'dry-run' ? 'skipped' : 'skipped',
      diff: createSimpleDiff(input.targetFile, existing, `${existing}\n\n${generatedBlock}`),
      warnings,
    };
  }

  const status: FileStatus = !exists ? 'created' : existing === nextContent ? 'unchanged' : 'updated';
  const diff = createSimpleDiff(input.targetFile, existing, nextContent);

  if (input.mode === 'write' && status !== 'unchanged') {
    await fs.mkdir(path.dirname(absoluteTarget), { recursive: true });
    await fs.writeFile(absoluteTarget, nextContent, 'utf-8');
  }

  return { status, diff, warnings };
}

async function runTypeCheck(repoRoot: string, command: string): Promise<BuildReport> {
  try {
    const { stdout, stderr } = await execAsync(command, { cwd: repoRoot });
    return {
      executed: true,
      success: true,
      command,
      summary: `${stdout}\n${stderr}`.trim().slice(-4000),
    };
  } catch (error: unknown) {
    const output = isExecError(error)
      ? `${error.stdout ?? ''}\n${error.stderr ?? ''}`.trim()
      : error instanceof Error
        ? error.message
        : 'Unknown build error';
    return {
      executed: true,
      success: false,
      command,
      summary: output.slice(-4000),
    };
  }
}

function unwrapResponseData(
  responses: Record<string, OpenApiResponse> | undefined,
  apiName: string,
  warnings: WarningItem[],
): OpenApiSchema | undefined {
  const response = responses?.['200'] ?? responses?.['201'] ?? responses?.['204'] ?? Object.entries(responses ?? {}).find(([code]) => code.startsWith('2'))?.[1];
  const schema = getJsonSchema(response?.content);
  if (!schema) {
    warnings.push({
      code: 'RESPONSE_SCHEMA_MISSING',
      message: `${apiName} has no JSON response schema, generated as Promise<void>`,
      api: apiName,
    });
    return undefined;
  }

  const resolved = schema;
  if (resolved.properties?.data) {
    return resolved.properties.data;
  }
  return resolved;
}

function schemaToTs(schema: OpenApiSchema, openApi: OpenApiDocument, warnings: WarningItem[]): string {
  const resolved = resolveRef(schema, openApi);

  if (resolved.allOf?.length) {
    return resolved.allOf.map((item) => schemaToTs(item, openApi, warnings)).join(' & ');
  }
  if (resolved.oneOf?.length) {
    return resolved.oneOf.map((item) => schemaToTs(item, openApi, warnings)).join(' | ');
  }
  if (resolved.anyOf?.length) {
    return resolved.anyOf.map((item) => schemaToTs(item, openApi, warnings)).join(' | ');
  }
  if (resolved.enum) {
    return renderEnum(resolved.enum);
  }

  const schemaType = Array.isArray(resolved.type) ? resolved.type.find((item) => item !== 'null') : resolved.type;
  const nullable = resolved.nullable || (Array.isArray(resolved.type) && resolved.type.includes('null'));
  let tsType: string;

  switch (schemaType) {
    case 'string':
      tsType = 'string';
      break;
    case 'integer':
    case 'number':
      tsType = 'number';
      break;
    case 'boolean':
      tsType = 'boolean';
      break;
    case 'array':
      tsType = `${resolved.items ? schemaToTs(resolved.items, openApi, warnings) : 'unknown'}[]`;
      break;
    case 'object':
      if (resolved.properties) {
        tsType = renderInlineObject(resolved, openApi, warnings);
      } else if (isObject(resolved.additionalProperties)) {
        tsType = `Record<string, ${schemaToTs(resolved.additionalProperties, openApi, warnings)}>`;
      } else {
        tsType = 'Record<string, unknown>';
      }
      break;
    case undefined:
      if (resolved.properties) {
        tsType = renderInlineObject(resolved, openApi, warnings);
      } else {
        tsType = 'unknown';
        warnings.push({
          code: 'SCHEMA_TYPE_MISSING',
          message: 'A schema is missing type, generated as unknown',
        });
      }
      break;
    default:
      tsType = 'unknown';
      warnings.push({
        code: 'UNSUPPORTED_SCHEMA_TYPE',
        message: `Unsupported schema type ${schemaType}, generated as unknown`,
      });
  }

  return nullable ? `${tsType} | null` : tsType;
}

function renderInlineObject(schema: OpenApiSchema, openApi: OpenApiDocument, warnings: WarningItem[]): string {
  const required = new Set(schema.required ?? []);
  const properties = schema.properties ?? {};
  const fields = Object.entries(properties).map(([name, propertySchema]) => {
    const optional = required.has(name) ? '' : '?';
    return `${safePropertyName(name)}${optional}: ${schemaToTs(propertySchema, openApi, warnings)}`;
  });
  return fields.length > 0 ? `{ ${fields.join('; ')} }` : 'Record<string, unknown>';
}

function resolveRef(schema: OpenApiSchema, openApi: OpenApiDocument): OpenApiSchema {
  if (!schema.$ref) {
    return schema;
  }

  const name = schema.$ref.split('/').pop();
  if (!name) return schema;
  return openApi.components?.schemas?.[name] ?? openApi.definitions?.[name] ?? schema;
}

function getJsonSchema(content: Record<string, { schema?: OpenApiSchema }> | undefined): OpenApiSchema | undefined {
  if (!content) return undefined;
  return (
    content['application/json']?.schema ??
    content['application/*+json']?.schema ??
    Object.values(content).find((item) => item.schema)?.schema
  );
}

function resolveResponseTypeName(functionName: string, schema: OpenApiSchema): string {
  const resolvedProperties = schema.properties ?? {};
  if ('access_token' in resolvedProperties && 'refresh_token' in resolvedProperties) {
    return 'TokenResp';
  }
  if ('items' in resolvedProperties && 'next_cursor' in resolvedProperties) {
    return `${pascalCase(functionName)}Resp`;
  }
  if (functionName === 'getMe' || functionName === 'updateMe') {
    return 'UserInfo';
  }
  if (functionName.toLowerCase().includes('task')) {
    return functionName === 'getTasks' ? 'ListTasksResp' : 'Task';
  }
  return `${pascalCase(functionName)}Resp`;
}

function resolveFunctionName(operationId: string | undefined, method: HttpMethod, frontendPath: string): string {
  if (operationId) {
    return camelCase(operationId);
  }

  const exactMap: Record<string, string> = {
    'post /auth/login': 'login',
    'post /auth/register': 'register',
    'post /auth/refresh': 'refreshToken',
    'post /auth/logout': 'logout',
    'get /users/me': 'getMe',
    'patch /users/me': 'updateMe',
    'put /users/me/password': 'changePassword',
    'get /tasks': 'getTasks',
    'post /tasks': 'createTask',
    'get /tasks/{id}': 'getTaskById',
    'patch /tasks/{id}': 'updateTask',
    'delete /tasks/{id}': 'deleteTask',
  };
  const exact = exactMap[`${method} ${frontendPath}`];
  if (exact) return exact;

  const segments = frontendPath
    .split('/')
    .filter(Boolean)
    .map((segment) => segment.replace(/[{}]/g, ''));
  return camelCase(`${method}-${segments.join('-')}`);
}

function resolveModuleName(frontendPath: string, fallback: ModuleConfig): string {
  const moduleConfig = Object.values(MODULES).find((item) =>
    item.pathPrefixes.some((prefix) => frontendPath.startsWith(prefix)),
  );
  return moduleConfig?.moduleName ?? fallback.moduleName;
}

function resolveTargetFile(args: ResolvedSyncArguments, moduleConfig: ModuleConfig): string {
  if (args.frontend.targetFile) {
    return args.frontend.targetFile;
  }
  if (args.scope.type === 'module') {
    return moduleConfig.targetFile;
  }
  return path.join(args.frontend.apiDir, `${moduleConfig.moduleName}.ts`);
}

function ensureSafeTarget(repoRoot: string, targetFile: string) {
  const absoluteApiDir = path.resolve(repoRoot, 'packages/frontend/src/api');
  const absoluteTarget = path.resolve(repoRoot, targetFile);
  if (!absoluteTarget.startsWith(absoluteApiDir + path.sep) && absoluteTarget !== absoluteApiDir) {
    throw new McpError(
      ErrorCode.InvalidParams,
      `Target file must stay under packages/frontend/src/api: ${targetFile}`,
    );
  }
}

function normalizeFrontendPath(inputPath: string, basePath: string): string {
  let normalized = inputPath.startsWith('/') ? inputPath : `/${inputPath}`;
  if (basePath && normalized.startsWith(basePath)) {
    normalized = normalized.slice(basePath.length) || '/';
  }
  return normalized.replace(/\/+/g, '/');
}

function extractGeneratedBlock(content: string, start: string, end: string): string | undefined {
  const startIndex = content.indexOf(start);
  const endIndex = content.indexOf(end);
  if (startIndex === -1 || endIndex === -1 || endIndex < startIndex) {
    return undefined;
  }
  return content.slice(startIndex, endIndex + end.length + (content[endIndex + end.length] === '\n' ? 1 : 0));
}

function createSimpleDiff(filePath: string, before: string, after: string): string | undefined {
  if (before === after) return undefined;
  const beforeLines = before.split('\n');
  const afterLines = after.split('\n');
  const output = [`--- ${filePath}`, `+++ ${filePath}`];
  output.push(`@@ before ${beforeLines.length} lines, after ${afterLines.length} lines @@`);
  output.push(...afterLines.slice(0, 200).map((line) => `+${line}`));
  if (afterLines.length > 200) {
    output.push(`+... ${afterLines.length - 200} more lines`);
  }
  return output.join('\n');
}

function buildNextSteps(args: ResolvedSyncArguments, fileStatus: FileStatus, buildReport: BuildReport): string[] {
  const steps: string[] = [];
  if (args.writeMode === 'dry-run') {
    steps.push('Review the dry-run diff, then rerun with writeMode=write to update files.');
  }
  if (fileStatus === 'skipped') {
    steps.push('Target file has no generated block. Set options.allowInsertGeneratedBlock=true or add the block manually.');
  }
  if (args.writeMode === 'write' && !buildReport.executed) {
    steps.push('Run pnpm --filter frontend build to verify generated TypeScript.');
  }
  if (buildReport.executed && buildReport.success === false) {
    steps.push('Fix the reported type-check errors before committing generated API code.');
  }
  return steps;
}

function failureReport(
  args: ResolvedSyncArguments,
  moduleConfig: ModuleConfig,
  code: string,
  message: string,
): SyncReport {
  return {
    success: false,
    mode: args.writeMode,
    scope: {
      type: args.scope.type,
      value: args.scope.value,
      matchedApiCount: 0,
    },
    target: {
      file: resolveTargetFile(args, moduleConfig),
      moduleName: moduleConfig.moduleName,
    },
    apis: [],
    types: [],
    files: [],
    warnings: [{ code, message }],
    nextSteps: ['Use a clear module, path-prefix, or operation-ids scope that exists in Apifox.'],
  };
}

async function findRepoRoot(start: string): Promise<string> {
  let current = path.resolve(start);
  for (;;) {
    const workspacePath = path.join(current, 'pnpm-workspace.yaml');
    const agentsPath = path.join(current, 'AGENTS.md');
    if (await exists(workspacePath) && await exists(agentsPath)) {
      return current;
    }
    const parent = path.dirname(current);
    if (parent === current) {
      throw new Error('Unable to locate repository root');
    }
    current = parent;
  }
}

async function exists(filePath: string): Promise<boolean> {
  try {
    await fs.access(filePath);
    return true;
  } catch {
    return false;
  }
}

function renderEnum(values: Array<string | number | boolean | null>): string {
  return values.map((value) => JSON.stringify(value)).join(' | ');
}

function isObjectSchema(schema: OpenApiSchema): boolean {
  return schema.type === 'object' || Boolean(schema.properties);
}

function asArray(value: string | string[]): string[] {
  return Array.isArray(value) ? value : [value];
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function isNodeError(error: unknown): error is NodeJS.ErrnoException {
  return error instanceof Error && 'code' in error;
}

function isExecError(error: unknown): error is Error & { stdout?: string; stderr?: string } {
  return error instanceof Error && ('stdout' in error || 'stderr' in error);
}

function safePropertyName(name: string): string {
  return /^[A-Za-z_$][A-Za-z0-9_$]*$/.test(name) ? name : JSON.stringify(name);
}

function camelCase(value: string): string {
  const words = value
    .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
    .split(/[^A-Za-z0-9]+/)
    .filter(Boolean);
  if (words.length === 0) return value;
  return words
    .map((word, index) => {
      const lower = word.toLowerCase();
      return index === 0 ? lower : lower.charAt(0).toUpperCase() + lower.slice(1);
    })
    .join('');
}

function pascalCase(value: string): string {
  const camel = camelCase(value);
  return camel.charAt(0).toUpperCase() + camel.slice(1);
}

const server = new ApifoxFrontendMcpServer();
server.run().catch((error: unknown) => {
  console.error(error);
  process.exit(1);
});
