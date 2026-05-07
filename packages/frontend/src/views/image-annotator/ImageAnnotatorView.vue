<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { ArrowLeft, Hand, PenTool, Undo2 } from "lucide-vue-next";
import paper from "paper";

const router = useRouter();
const { t } = useI18n();

type ShapeKind = "rectangle" | "ellipse" | "triangle" | "star" | "heart";
type ToolMode = "select" | ShapeKind;
interface AnnotationSnapshotItem {
  id: string;
  kind: ShapeKind;
  bounds: { x: number; y: number; width: number; height: number };
  style: {
    fillEnabled: boolean;
    fillColor: string;
    edgeColor: string;
    edgeWidth: number;
    opacity: number;
    dashArray: number[];
  };
}

const containerRef = ref<HTMLElement | null>(null);
const canvasRef = ref<HTMLCanvasElement | null>(null);
const fileInputRef = ref<HTMLInputElement | null>(null);

const activeTool = ref<ToolMode>("star");
const isBusy = ref(false);
const hasImage = ref(false);
const errorMessage = ref("");

const fillColor = ref("#00f3ff");
const fillEnabled = ref(false);
const strokeColor = ref("#ffffff");
const strokeWidth = ref(2);
const opacity = ref(0.95);
const isDashed = ref(false);
const dashLength = ref(8);
const dashGap = ref(6);
const annotationSize = ref(96);

const activeItemId = ref<string>("");

let scope: paper.PaperScope | null = null;
let drawingTool: paper.Tool | null = null;
let backgroundLayer: paper.Layer | null = null;
let annotationLayer: paper.Layer | null = null;
let backgroundRaster: paper.Raster | null = null;
let dragItem: paper.Item | null = null;
let creatingItem: paper.Item | null = null;
let createStartPoint: paper.Point | null = null;
let selectedItem: paper.Item | null = null;
let resizeItem: paper.Item | null = null;
let resizeCenter: paper.Point | null = null;
let lastResizeDistance = 0;
let isPanning = false;
let panStartPoint: paper.Point | null = null;
let panStartCenter: paper.Point | null = null;
let panStartClientX = 0;
let panStartClientY = 0;
let mutatedDuringPointer = false;
let uid = 1;
let historyStack: string[] = [];
const historyIndex = ref(-1);
let suppressHistory = false;
const syncSelectionState = ref(false);
let styleCommitTimer: number | null = null;
const MIN_CANVAS_ZOOM = 1;
const MAX_CANVAS_ZOOM = 4;

const canExport = computed(() => hasImage.value && Boolean(scope));
const canUndo = computed(() => historyIndex.value > 0);
const toolOptions = computed<Array<{ label: string; value: ToolMode }>>(() => [
  { label: t("annotator.tools.rectangle"), value: "rectangle" },
  { label: t("annotator.tools.ellipse"), value: "ellipse" },
  { label: t("annotator.tools.triangle"), value: "triangle" },
  { label: t("annotator.tools.star"), value: "star" },
  { label: t("annotator.tools.heart"), value: "heart" },
]);

onMounted(async () => {
  await nextTick();
  setupPaper();
  window.addEventListener("resize", handleResize);
  canvasRef.value?.addEventListener("wheel", handleWheel, { passive: false });
  document.addEventListener("wheel", preventBrowserZoom, { passive: false });
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", handleResize);
  canvasRef.value?.removeEventListener("wheel", handleWheel);
  document.removeEventListener("wheel", preventBrowserZoom);
  drawingTool?.remove();
  scope?.project.clear();
  if (styleCommitTimer !== null) {
    window.clearTimeout(styleCommitTimer);
    styleCommitTimer = null;
  }
});

watch([fillEnabled, fillColor, strokeColor, strokeWidth, opacity, isDashed, dashLength, dashGap], () => {
  if (syncSelectionState.value) {
    return;
  }
  applyStylesToSelected();
  if (selectedItem) {
    scheduleHistoryCommit();
  }
});

watch(annotationSize, (nextSize) => {
  if (syncSelectionState.value) {
    return;
  }
  applySizeToSelected(nextSize);
  if (selectedItem) {
    scheduleHistoryCommit();
  }
});

watch(activeTool, (mode) => {
  const canvas = canvasRef.value;
  if (!canvas) {
    return;
  }
  canvas.style.cursor = mode === "select" ? "grab" : "crosshair";
});

function setupPaper() {
  const canvas = canvasRef.value;
  const container = containerRef.value;
  if (!canvas || !container) {
    return;
  }

  scope = new paper.PaperScope();
  resizeCanvas();
  scope.setup(canvas);

  backgroundLayer = new scope.Layer();
  annotationLayer = new scope.Layer();
  annotationLayer.activate();

  drawingTool = new scope.Tool();
  drawingTool.onMouseDown = (event: paper.ToolEvent) => {
    const nativeEvent = (event as unknown as { event?: MouseEvent }).event;
    handleMouseDown(event.point, nativeEvent);
  };
  drawingTool.onMouseDrag = (event: paper.ToolEvent) => {
    const nativeEvent = (event as unknown as { event?: MouseEvent }).event;
    handleMouseDrag(event.point, event.delta, nativeEvent);
  };
  drawingTool.onMouseMove = (event: paper.ToolEvent) => {
    handleMouseMove(event.point);
  };
  drawingTool.onMouseUp = (event: paper.ToolEvent) => {
    handleMouseUp(event.point);
  };

  if (canvasRef.value) {
    canvasRef.value.style.cursor = activeTool.value === "select" ? "grab" : "crosshair";
  }
}

function resizeCanvas() {
  const canvas = canvasRef.value;
  const container = containerRef.value;
  if (!canvas || !container) {
    return;
  }

  const width = Math.max(container.clientWidth, 320);
  const height = Math.max(container.clientHeight, 320);
  canvas.width = width;
  canvas.height = height;
}

function handleResize() {
  resizeCanvas();
  if (!scope) {
    return;
  }

  scope.view.viewSize = new scope.Size(canvasRef.value?.width || 0, canvasRef.value?.height || 0);
  fitBackgroundToView();
}

function triggerUpload() {
  fileInputRef.value?.click();
}

function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (!file) {
    return;
  }

  if (!file.type.startsWith("image/")) {
    errorMessage.value = t("annotator.errors.onlyImage");
    return;
  }

  errorMessage.value = "";
  isBusy.value = true;

  const reader = new FileReader();
  reader.onload = () => {
    const result = reader.result;
    if (typeof result !== "string") {
      errorMessage.value = t("annotator.errors.readFailed");
      isBusy.value = false;
      return;
    }
    loadBackgroundImage(result);
  };
  reader.onerror = () => {
    errorMessage.value = t("annotator.errors.readFailed");
    isBusy.value = false;
  };
  reader.readAsDataURL(file);
  target.value = "";
}

function loadBackgroundImage(dataUrl: string) {
  if (!scope || !backgroundLayer || !annotationLayer) {
    isBusy.value = false;
    return;
  }

  backgroundLayer.activate();
  backgroundRaster?.remove();

  const raster = new scope.Raster(dataUrl);
  raster.onLoad = () => {
    backgroundRaster = raster;
    hasImage.value = true;
    errorMessage.value = "";
    fitBackgroundToView();
    annotationLayer?.activate();
    clearAnnotations();
    resetHistory();
    pushHistorySnapshot();
    isBusy.value = false;
  };
  raster.onError = () => {
    errorMessage.value = t("annotator.errors.loadFailed");
    isBusy.value = false;
  };
}

function fitBackgroundToView() {
  if (!scope || !backgroundRaster) {
    return;
  }

  const margin = 20;
  const viewBounds = scope.view.bounds;
  const fitRect = new scope.Rectangle(
    viewBounds.x + margin,
    viewBounds.y + margin,
    Math.max(viewBounds.width - margin * 2, 120),
    Math.max(viewBounds.height - margin * 2, 120),
  );

  backgroundRaster.fitBounds(fitRect, true);
  backgroundRaster.sendToBack();
}

function clearAnnotations() {
  if (!annotationLayer) {
    return;
  }

  annotationLayer.removeChildren();
  selectedItem = null;
  dragItem = null;
  resizeItem = null;
  resizeCenter = null;
  lastResizeDistance = 0;
  isPanning = false;
  panStartPoint = null;
  panStartCenter = null;
  panStartClientX = 0;
  panStartClientY = 0;
  mutatedDuringPointer = false;
  creatingItem = null;
  createStartPoint = null;
  activeItemId.value = "";
}

function resetHistory() {
  historyStack = [];
  historyIndex.value = -1;
}

function scheduleHistoryCommit() {
  if (suppressHistory) {
    return;
  }
  if (styleCommitTimer !== null) {
    window.clearTimeout(styleCommitTimer);
  }
  styleCommitTimer = window.setTimeout(() => {
    styleCommitTimer = null;
    pushHistorySnapshot();
  }, 180);
}

function pushHistorySnapshot() {
  if (!annotationLayer || suppressHistory) {
    return;
  }
  const snapshot = serializeAnnotations();
  if (historyIndex.value >= 0 && historyStack[historyIndex.value] === snapshot) {
    return;
  }
  if (historyIndex.value < historyStack.length - 1) {
    historyStack = historyStack.slice(0, historyIndex.value + 1);
  }
  historyStack.push(snapshot);
  historyIndex.value = historyStack.length - 1;
}

function serializeAnnotations(): string {
  if (!annotationLayer) {
    return "[]";
  }
  const items: AnnotationSnapshotItem[] = [];
  annotationLayer.children.forEach((item) => {
    const kind = item.data.kind as ShapeKind | undefined;
    if (!kind) {
      return;
    }
    items.push({
      id: String(item.data.id || ""),
      kind,
      bounds: {
        x: item.bounds.x,
        y: item.bounds.y,
        width: item.bounds.width,
        height: item.bounds.height,
      },
      style: {
        fillEnabled: item.fillColor !== null,
        fillColor: toHexColor(item.fillColor, "#00f3ff"),
        edgeColor: toHexColor(item.strokeColor, "#ffffff"),
        edgeWidth: item.strokeWidth || 1,
        opacity: item.opacity,
        dashArray: (item.dashArray || []).map((v) => Number(v)),
      },
    });
  });
  return JSON.stringify(items);
}

function restoreAnnotationsFromSnapshot(snapshot: string) {
  if (!scope || !annotationLayer) {
    return;
  }
  const paperScope = scope;
  const parsed = JSON.parse(snapshot) as AnnotationSnapshotItem[];
  annotationLayer.removeChildren();
  deselectItem();

  let maxId = uid;
  parsed.forEach((entry) => {
    const rect = new paperScope.Rectangle(entry.bounds.x, entry.bounds.y, entry.bounds.width, entry.bounds.height);
    const item = createShapeByBounds(entry.kind, rect);
    if (!item) {
      return;
    }
    item.data = { id: entry.id, kind: entry.kind };
    item.fillColor = entry.style.fillEnabled ? new paper.Color(entry.style.fillColor) : null;
    item.strokeColor = new paper.Color(entry.style.edgeColor);
    item.strokeWidth = entry.style.edgeWidth;
    item.opacity = entry.style.opacity;
    item.dashArray = [...entry.style.dashArray];
    constrainItemWithinImage(item);

    const idNum = Number(String(entry.id).replace("anno-", ""));
    if (Number.isFinite(idNum)) {
      maxId = Math.max(maxId, idNum + 1);
    }
  });
  uid = maxId;
}

function undoHistory() {
  if (historyIndex.value <= 0) {
    return;
  }
  historyIndex.value -= 1;
  suppressHistory = true;
  restoreAnnotationsFromSnapshot(historyStack[historyIndex.value]);
  suppressHistory = false;
}

function resolveTopItemByPoint(point: paper.Point): paper.Item | null {
  if (!annotationLayer) {
    return null;
  }

  for (let i = annotationLayer.children.length - 1; i >= 0; i -= 1) {
    const item = annotationLayer.children[i];
    const isInside = item.contains(point);
    const edgeHit = item.hitTest(point, {
      fill: true,
      stroke: true,
      segments: false,
      tolerance: 8,
    });
    if (isInside || Boolean(edgeHit)) {
      return item;
    }
  }

  return null;
}

function handleMouseDown(point: paper.Point, nativeEvent?: MouseEvent) {
  if (!scope || !annotationLayer) {
    return;
  }
  mutatedDuringPointer = false;

  if (!hasImage.value) {
    if (activeTool.value !== "select") {
      errorMessage.value = t("annotator.errors.uploadFirst");
    }
    return;
  }

  if (activeTool.value === "select") {
    const edgeHit = scope.project.hitTest(point, {
      fill: false,
      stroke: true,
      segments: true,
      tolerance: 8,
    });

    if (edgeHit?.item && edgeHit.item.layer === annotationLayer && (edgeHit.type === "stroke" || edgeHit.type === "segment" || edgeHit.type === "curve")) {
      selectItem(edgeHit.item);
      resizeItem = edgeHit.item;
      resizeCenter = edgeHit.item.position.clone();
      lastResizeDistance = Math.max(point.getDistance(resizeCenter), 1);
      if (canvasRef.value) {
        canvasRef.value.style.cursor = "nwse-resize";
      }
      return;
    }

    const bodyHit = resolveTopItemByPoint(point);
    if (bodyHit) {
      selectItem(bodyHit);
      dragItem = bodyHit;
      if (canvasRef.value) {
        canvasRef.value.style.cursor = "grabbing";
      }
      return;
    }

    deselectItem();
    isPanning = canPanView();
    if (isPanning) {
      panStartPoint = point.clone();
      panStartCenter = scope.view.center.clone();
      panStartClientX = nativeEvent?.clientX ?? 0;
      panStartClientY = nativeEvent?.clientY ?? 0;
    } else {
      panStartPoint = null;
      panStartCenter = null;
      panStartClientX = 0;
      panStartClientY = 0;
    }
    if (canvasRef.value) {
      canvasRef.value.style.cursor = isPanning ? "grabbing" : "grab";
    }
    return;
  }

  errorMessage.value = "";
  if (!isPointInsideImage(point)) {
    return;
  }
  const startPoint = clampPointToImage(point);
  createStartPoint = startPoint;
  const startRect = new scope.Rectangle(startPoint, startPoint);
  creatingItem = createShapeByBounds(activeTool.value, startRect);
  if (creatingItem) {
    selectItem(creatingItem);
  }
}

function handleMouseDrag(point: paper.Point, delta: paper.Point, nativeEvent?: MouseEvent) {
  if (!scope) {
    return;
  }

  if (creatingItem && createStartPoint && activeTool.value !== "select") {
    const nextRect = new scope.Rectangle(createStartPoint, clampPointToImage(point));
    replaceCreatingShape(nextRect);
    mutatedDuringPointer = true;
    return;
  }

  if (!dragItem || activeTool.value !== "select") {
    if (isPanning) {
      panViewFromStart(point, nativeEvent);
      return;
    }
    if (resizeItem && resizeCenter) {
      const nextDistance = Math.max(point.getDistance(resizeCenter), 1);
      const ratio = nextDistance / Math.max(lastResizeDistance, 1);
      if (Number.isFinite(ratio) && ratio > 0) {
        resizeItem.scale(ratio, resizeCenter);
        constrainItemWithinImage(resizeItem);
        lastResizeDistance = nextDistance;
        mutatedDuringPointer = true;
      }
    }
    return;
  }
  dragItem.position = dragItem.position.add(delta);
  constrainItemWithinImage(dragItem);
  mutatedDuringPointer = true;
}

function handleMouseMove(point: paper.Point) {
  const canvas = canvasRef.value;
  if (!scope || !canvas || !annotationLayer) {
    return;
  }

  if (activeTool.value !== "select") {
    canvas.style.cursor = "crosshair";
    return;
  }

  const edgeHit = scope.project.hitTest(point, {
    fill: false,
    stroke: true,
    segments: true,
    tolerance: 8,
  });
  if (edgeHit?.item && edgeHit.item.layer === annotationLayer && (edgeHit.type === "stroke" || edgeHit.type === "segment" || edgeHit.type === "curve")) {
    canvas.style.cursor = "nwse-resize";
    return;
  }

  const bodyHit = resolveTopItemByPoint(point);
  if (bodyHit) {
    canvas.style.cursor = "grab";
    return;
  }

  canvas.style.cursor = "grab";
}

function handleMouseUp(point: paper.Point) {
  const canvas = canvasRef.value;
  if (!scope) {
    dragItem = null;
    resizeItem = null;
    resizeCenter = null;
    lastResizeDistance = 0;
    isPanning = false;
    panStartPoint = null;
    panStartCenter = null;
    panStartClientX = 0;
    panStartClientY = 0;
    if (canvas) {
      canvas.style.cursor = activeTool.value === "select" ? "grab" : "crosshair";
    }
    return;
  }

  if (creatingItem && createStartPoint && activeTool.value !== "select") {
    const finalRect = new scope.Rectangle(createStartPoint, clampPointToImage(point));
    const isTiny = Math.abs(finalRect.width) < 6 && Math.abs(finalRect.height) < 6;
    if (isTiny) {
      creatingItem.remove();
      creatingItem = null;
      createStartPoint = null;
      deselectItem();
      mutatedDuringPointer = false;
      return;
    }

    replaceCreatingShape(finalRect);
    creatingItem = null;
    createStartPoint = null;
  }

  if (mutatedDuringPointer) {
    pushHistorySnapshot();
  }
  mutatedDuringPointer = false;

  dragItem = null;
  resizeItem = null;
  resizeCenter = null;
  lastResizeDistance = 0;
  isPanning = false;
  panStartPoint = null;
  panStartCenter = null;
  panStartClientX = 0;
  panStartClientY = 0;
  if (canvas) {
    canvas.style.cursor = activeTool.value === "select" ? "grab" : "crosshair";
  }
}

function handleWheel(event: WheelEvent) {
  if (!scope) {
    return;
  }
  event.preventDefault();
  event.stopPropagation();
  if (activeTool.value !== "select") {
    return;
  }

  const zoomFactor = event.deltaY < 0 ? 1.06 : 0.94;
  const nextZoom = Math.min(MAX_CANVAS_ZOOM, Math.max(MIN_CANVAS_ZOOM, scope.view.zoom * zoomFactor));
  if (Math.abs(nextZoom - scope.view.zoom) < 0.0001) {
    return;
  }
  const point = scope.view.viewToProject(new scope.Point(event.offsetX, event.offsetY));
  scope.view.scale(nextZoom / scope.view.zoom, point);
  clampViewCenter();
}

function preventBrowserZoom(event: WheelEvent) {
  if (!event.ctrlKey) {
    return;
  }
  const canvas = canvasRef.value;
  if (!canvas) {
    return;
  }
  const target = event.target as Node | null;
  if (target && canvas.contains(target)) {
    event.preventDefault();
  }
}

function canPanView(): boolean {
  if (!scope || !backgroundRaster) {
    return false;
  }
  return scope.view.zoom > 1.001;
}

function getImageBounds(): paper.Rectangle | null {
  if (!backgroundRaster) {
    return null;
  }
  return backgroundRaster.bounds.clone();
}

function isPointInsideImage(point: paper.Point): boolean {
  const imageBounds = getImageBounds();
  if (!imageBounds) {
    return false;
  }
  return (
    point.x >= imageBounds.left &&
    point.x <= imageBounds.right &&
    point.y >= imageBounds.top &&
    point.y <= imageBounds.bottom
  );
}

function clampPointToImage(point: paper.Point): paper.Point {
  const imageBounds = getImageBounds();
  if (!scope || !imageBounds) {
    return point;
  }
  const clampedX = Math.min(imageBounds.right, Math.max(imageBounds.left, point.x));
  const clampedY = Math.min(imageBounds.bottom, Math.max(imageBounds.top, point.y));
  return new scope.Point(clampedX, clampedY);
}

function constrainItemWithinImage(item: paper.Item) {
  if (!scope) {
    return;
  }
  const imageBounds = getImageBounds();
  if (!imageBounds) {
    return;
  }

  if (item.bounds.width > imageBounds.width || item.bounds.height > imageBounds.height) {
    item.fitBounds(imageBounds, true);
  }

  const currentBounds = item.bounds;
  let deltaX = 0;
  let deltaY = 0;

  if (currentBounds.left < imageBounds.left) {
    deltaX = imageBounds.left - currentBounds.left;
  } else if (currentBounds.right > imageBounds.right) {
    deltaX = imageBounds.right - currentBounds.right;
  }

  if (currentBounds.top < imageBounds.top) {
    deltaY = imageBounds.top - currentBounds.top;
  } else if (currentBounds.bottom > imageBounds.bottom) {
    deltaY = imageBounds.bottom - currentBounds.bottom;
  }

  if (deltaX !== 0 || deltaY !== 0) {
    item.position = item.position.add(new scope.Point(deltaX, deltaY));
  }
}

function panViewFromStart(currentPoint: paper.Point, nativeEvent?: MouseEvent) {
  if (!scope || !backgroundRaster || !panStartPoint || !panStartCenter) {
    return;
  }
  if (!canPanView()) {
    return;
  }
  const panSpeed = 1.15;
  if (nativeEvent && panStartClientX !== 0 && panStartClientY !== 0) {
    const dx = (nativeEvent.clientX - panStartClientX) / scope.view.zoom;
    const dy = (nativeEvent.clientY - panStartClientY) / scope.view.zoom;
    const dragVector = new scope.Point(dx, dy).multiply(panSpeed);
    scope.view.center = panStartCenter.subtract(dragVector);
  } else {
    const dragVector = currentPoint.subtract(panStartPoint).multiply(panSpeed);
    scope.view.center = panStartCenter.subtract(dragVector);
  }
  clampViewCenter();
}

function clampViewCenter() {
  if (!scope || !backgroundRaster) {
    return;
  }

  const imageBounds = backgroundRaster.bounds;
  const viewBounds = scope.view.bounds;
  const halfW = viewBounds.width / 2;
  const halfH = viewBounds.height / 2;

  let nextX = scope.view.center.x;
  let nextY = scope.view.center.y;

  if (imageBounds.width <= viewBounds.width) {
    nextX = imageBounds.center.x;
  } else {
    const minX = imageBounds.left + halfW;
    const maxX = imageBounds.right - halfW;
    nextX = Math.min(maxX, Math.max(minX, nextX));
  }

  if (imageBounds.height <= viewBounds.height) {
    nextY = imageBounds.center.y;
  } else {
    const minY = imageBounds.top + halfH;
    const maxY = imageBounds.bottom - halfH;
    nextY = Math.min(maxY, Math.max(minY, nextY));
  }

  scope.view.center = new scope.Point(nextX, nextY);
}

function replaceCreatingShape(bounds: paper.Rectangle) {
  if (!scope || !creatingItem || activeTool.value === "select") {
    return;
  }

  const cachedData = creatingItem.data;
  const keepSelected = creatingItem.selected;
  creatingItem.remove();

  const next = createShapeByBounds(activeTool.value, bounds);
  if (!next) {
    creatingItem = null;
    return;
  }

  next.data = cachedData;
  next.selected = keepSelected;
  creatingItem = next;
  selectedItem = next;
}

function createShapeByBounds(shape: ShapeKind, rawBounds: paper.Rectangle): paper.Item | null {
  if (!scope || !annotationLayer) {
    return null;
  }

  annotationLayer.activate();
  const bounds = normalizeBounds(rawBounds);
  const center = bounds.center;
  let item: paper.Item;

  if (shape === "rectangle") {
    item = new scope.Path.Rectangle({
      rectangle: bounds,
    });
  } else if (shape === "ellipse") {
    item = new scope.Path.Ellipse({ rectangle: bounds });
  } else if (shape === "triangle") {
    const top = new scope.Point(center.x, bounds.top);
    const left = new scope.Point(bounds.left, bounds.bottom);
    const right = new scope.Point(bounds.right, bounds.bottom);
    const triangle = new scope.Path([top, right, left]);
    triangle.closed = true;
    item = triangle;
  } else if (shape === "star") {
    const outerRadius = Math.max(bounds.width, bounds.height) / 2;
    const innerRadius = outerRadius * 0.45;
    const star = new scope.Path.Star(center, 5, innerRadius, outerRadius);
    star.fitBounds(bounds, true);
    item = star;
  } else {
    const heart = createHeartPath(center, Math.max(bounds.width, bounds.height));
    heart.fitBounds(bounds, true);
    item = heart;
  }

  if (!item.data.id) {
    item.data = {
      id: `anno-${uid++}`,
      kind: shape,
    };
  }

  applyStylesToItem(item);
  constrainItemWithinImage(item);
  return item;
}

function normalizeBounds(rawBounds: paper.Rectangle): paper.Rectangle {
  if (!scope) {
    throw new Error("Paper scope is not initialized");
  }
  const minSide = 8;
  const x = Math.min(rawBounds.x, rawBounds.x + rawBounds.width);
  const y = Math.min(rawBounds.y, rawBounds.y + rawBounds.height);
  const width = Math.max(Math.abs(rawBounds.width), minSide);
  const height = Math.max(Math.abs(rawBounds.height), minSide);
  return new scope.Rectangle(x, y, width, height);
}

function createHeartPath(center: paper.Point, size: number): paper.Path {
  if (!scope) {
    throw new Error("Paper scope is not initialized");
  }

  const base = new scope.Path("M 0 18 C -38 -10 -30 -48 0 -26 C 30 -48 38 -10 0 18 Z");
  base.closed = true;
  const scaleRatio = size / Math.max(base.bounds.width, base.bounds.height);
  base.scale(scaleRatio);
  base.position = center;
  return base;
}

function selectItem(item: paper.Item) {
  if (selectedItem) {
    selectedItem.selected = false;
  }
  selectedItem = item;
  selectedItem.selected = true;

  const id = typeof selectedItem.data.id === "string" ? selectedItem.data.id : "";
  activeItemId.value = id;

  syncSelectionState.value = true;
  fillColor.value = toHexColor(selectedItem.fillColor, "#00f3ff");
  fillEnabled.value = selectedItem.fillColor !== null;
  strokeColor.value = toHexColor(selectedItem.strokeColor, "#ffffff");
  strokeWidth.value = selectedItem.strokeWidth || 1;
  opacity.value = selectedItem.opacity;
  const dashArray = selectedItem.dashArray || [];
  isDashed.value = dashArray.length > 0;
  dashLength.value = dashArray.length >= 1 ? dashArray[0] : 8;
  dashGap.value = dashArray.length >= 2 ? dashArray[1] : 6;
  annotationSize.value = Math.max(selectedItem.bounds.width, selectedItem.bounds.height);
  void nextTick(() => {
    syncSelectionState.value = false;
  });
}

function deselectItem() {
  if (selectedItem) {
    selectedItem.selected = false;
  }
  selectedItem = null;
  activeItemId.value = "";
}

function applyStylesToSelected() {
  if (!selectedItem) {
    return;
  }
  applyStylesToItem(selectedItem);
}

function applyStylesToItem(item: paper.Item) {
  item.fillColor = fillEnabled.value ? new paper.Color(fillColor.value) : null;
  item.strokeColor = new paper.Color(strokeColor.value);
  item.strokeWidth = strokeWidth.value;
  item.opacity = opacity.value;
  item.dashArray = isDashed.value ? [dashLength.value, dashGap.value] : [];
}

function applySizeToSelected(nextSize: number) {
  if (!selectedItem) {
    return;
  }

  const current = Math.max(selectedItem.bounds.width, selectedItem.bounds.height);
  if (current <= 0) {
    return;
  }

  const ratio = nextSize / current;
  if (!Number.isFinite(ratio) || ratio <= 0) {
    return;
  }

  selectedItem.scale(ratio);
  constrainItemWithinImage(selectedItem);
}

function removeSelected() {
  if (!selectedItem) {
    return;
  }
  selectedItem.remove();
  selectedItem = null;
  activeItemId.value = "";
  pushHistorySnapshot();
}

function exportImage() {
  const canvas = canvasRef.value;
  if (!canvas || !canExport.value) {
    return;
  }

  const dataUrl = canvas.toDataURL("image/png");
  const link = document.createElement("a");
  link.href = dataUrl;
  link.download = `annotated_${Date.now()}.png`;
  link.click();
}

function toHexColor(color: paper.Color | null, fallback: string): string {
  if (!color) {
    return fallback;
  }
  return color.toCSS(true);
}

function backToTasks() {
  router.push("/tasks");
}
</script>

<template>
  <main class="min-h-screen bg-[#050a0f] text-white">
    <section class="relative mx-auto max-w-[1400px] px-4 py-6 md:px-6 md:py-8">
      <button
        class="absolute left-4 top-6 inline-flex items-center gap-1.5 rounded-full border border-white/20 px-3 py-1.5 text-sm text-white/85 transition-colors hover:border-neon hover:text-neon md:left-6 md:top-8"
        @click="backToTasks"
      >
        <ArrowLeft class="h-4 w-4" />
        <span class="hidden sm:inline">{{ t("tasks.backToList") }}</span>
      </button>

      <header class="mb-6 flex items-center justify-center">
        <h1 class="text-neon text-center text-3xl font-semibold tracking-[0.18em] drop-shadow-[0_0_18px_var(--neon-glow)] md:text-4xl">
          {{ t("annotator.title") }}
        </h1>
      </header>

      <div class="mb-4 flex flex-wrap items-center justify-end gap-2 rounded-xl border border-white/10 bg-white/5 px-4 py-3">
        <input
          ref="fileInputRef"
          class="hidden"
          type="file"
          accept="image/png,image/jpeg,image/webp,image/jpg"
          @change="handleFileChange"
        />
        <button
          class="rounded-full border border-white/20 px-3 py-1.5 text-sm text-white/85 transition-colors hover:border-neon hover:text-neon"
          :disabled="isBusy"
          @click="triggerUpload"
        >
          {{ isBusy ? t("common.loading") : t("annotator.upload") }}
        </button>
        <button
          class="rounded-full border border-neon/70 bg-neon/10 px-3 py-1.5 text-sm text-neon transition-colors hover:bg-neon/20 disabled:cursor-not-allowed disabled:opacity-50"
          :disabled="!canUndo"
          @click="undoHistory"
        >
          <span class="inline-flex items-center gap-1.5">
            <Undo2 class="h-3.5 w-3.5" />
            {{ t("annotator.undo") }}
          </span>
        </button>
        <button
          class="rounded-full border border-neon/70 bg-neon/10 px-3 py-1.5 text-sm text-neon transition-colors hover:bg-neon/20 disabled:cursor-not-allowed disabled:opacity-50"
          :disabled="!canExport"
          @click="exportImage"
        >
          {{ t("annotator.download") }}
        </button>
      </div>

      <p
        v-if="errorMessage"
        class="mb-3 text-sm text-rose-300"
      >
        {{ errorMessage }}
      </p>

      <section class="grid grid-cols-1 gap-3 md:h-[calc(100vh-220px)] md:grid-cols-[220px_minmax(0,1fr)_280px] md:p-0">
      <aside class="order-2 rounded-2xl border border-white/10 bg-[#0a121b]/90 p-3 md:order-1 md:overflow-y-auto">
        <p class="mb-2 text-sm text-white/60">{{ t("annotator.mode") }}</p>
        <div class="mb-3 grid grid-cols-2 gap-2">
          <button
            class="inline-flex items-center justify-center gap-1 rounded-xl border px-2 py-2 text-sm transition-colors"
            :class="activeTool === 'select' ? 'border-neon bg-neon/20 text-neon' : 'border-white/15 text-white/80 hover:border-white/30'"
            @click="activeTool = 'select'"
          >
            <Hand class="h-4 w-4" />
            {{ t("annotator.handMode") }}
          </button>
          <button
            class="inline-flex items-center justify-center gap-1 rounded-xl border px-2 py-2 text-sm transition-colors"
            :class="activeTool !== 'select' ? 'border-neon bg-neon/20 text-neon' : 'border-white/15 text-white/80 hover:border-white/30'"
            @click="activeTool = 'star'"
          >
            <PenTool class="h-4 w-4" />
            {{ t("annotator.annotateMode") }}
          </button>
        </div>

        <p class="mb-2 text-sm text-white/60">{{ t("annotator.drawTools") }}</p>
        <div class="grid grid-cols-2 gap-2">
          <button
            v-for="tool in toolOptions"
            :key="tool.value"
            class="rounded-xl border px-2 py-2 text-sm transition-colors"
            :disabled="activeTool === 'select'"
            :class="activeTool === tool.value ? 'border-neon bg-neon/20 text-neon' : 'border-white/15 text-white/80 hover:border-white/30 disabled:cursor-not-allowed disabled:opacity-50'"
            @click="activeTool = tool.value"
          >
            {{ tool.label }}
          </button>
        </div>
      </aside>

      <div
        ref="containerRef"
        class="relative order-1 min-h-[52vh] overflow-hidden rounded-2xl border border-white/10 bg-[#05080d] md:order-2 md:min-h-0"
      >
        <canvas
          ref="canvasRef"
          class="block h-full w-full"
        />
        <div
          v-if="!hasImage"
          class="pointer-events-none absolute inset-0 flex items-center justify-center text-sm text-white/45"
        >
          {{ t("annotator.uploadHint") }}
        </div>
      </div>

      <aside class="order-3 rounded-2xl border border-white/10 bg-[#0a121b]/90 p-3 md:overflow-y-auto">
        <p class="mb-2 text-sm text-white/60">{{ t("annotator.properties") }}</p>
        <p class="mb-3 text-xs text-white/40">
          {{ t("annotator.currentId") }}: {{ activeItemId || t("annotator.noneSelected") }}
        </p>

        <div class="space-y-3">
          <label class="flex items-center gap-2 text-sm text-white/85">
            <input
              v-model="fillEnabled"
              type="checkbox"
            />
            {{ t("annotator.fillEnabled") }}
          </label>

          <label class="block text-sm text-white/85">
            {{ t("annotator.fillColor") }}
            <input
              v-model="fillColor"
              class="mt-1 h-9 w-full rounded-lg border border-white/15 bg-transparent px-2"
              type="color"
              :disabled="!fillEnabled"
            />
          </label>

          <label class="block text-sm text-white/85">
            {{ t("annotator.edgeColor") }}
            <input
              v-model="strokeColor"
              class="mt-1 h-9 w-full rounded-lg border border-white/15 bg-transparent px-2"
              type="color"
              :disabled="!activeItemId"
            />
          </label>

          <label class="block text-sm text-white/85">
            {{ t("annotator.edgeWidth") }}: {{ strokeWidth }}
            <input
              v-model.number="strokeWidth"
              class="mt-1 w-full"
              type="range"
              min="1"
              max="14"
              step="1"
              :disabled="!activeItemId"
            />
          </label>

          <label class="block text-sm text-white/85">
            {{ t("annotator.opacity") }}: {{ opacity.toFixed(2) }}
            <input
              v-model.number="opacity"
              class="mt-1 w-full"
              type="range"
              min="0.1"
              max="1"
              step="0.05"
              :disabled="!activeItemId"
            />
          </label>

          <label class="block text-sm text-white/85">
            {{ t("annotator.size") }}: {{ Math.round(annotationSize) }}
            <input
              v-model.number="annotationSize"
              class="mt-1 w-full"
              type="range"
              min="24"
              max="260"
              step="2"
              :disabled="!activeItemId"
            />
          </label>

          <label class="flex items-center gap-2 text-sm text-white/85">
            <input
              v-model="isDashed"
              type="checkbox"
              :disabled="!activeItemId"
            />
            {{ t("annotator.dashedEdge") }}
          </label>

          <label class="block text-sm text-white/85">
            {{ t("annotator.dashLength") }}: {{ dashLength }}
            <input
              v-model.number="dashLength"
              class="mt-1 w-full"
              type="range"
              min="2"
              max="24"
              step="1"
              :disabled="!activeItemId || !isDashed"
            />
          </label>

          <label class="block text-sm text-white/85">
            {{ t("annotator.dashGap") }}: {{ dashGap }}
            <input
              v-model.number="dashGap"
              class="mt-1 w-full"
              type="range"
              min="2"
              max="24"
              step="1"
              :disabled="!activeItemId || !isDashed"
            />
          </label>

          <button
            class="mt-2 w-full rounded-xl border border-rose-400/40 bg-rose-500/10 px-3 py-2 text-sm text-rose-200 transition-colors hover:bg-rose-500/20 disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="!activeItemId"
            @click="removeSelected"
          >
            {{ t("annotator.removeSelected") }}
          </button>
        </div>
      </aside>
      </section>
    </section>
  </main>
</template>
