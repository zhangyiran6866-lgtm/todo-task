<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { ArrowLeft, ChevronDown } from "lucide-vue-next";
import * as THREE from "three";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import { PLYLoader } from "three/examples/jsm/loaders/PLYLoader.js";

const router = useRouter();
const { t } = useI18n();

const canvasContainerRef = ref<HTMLElement | null>(null);
const renderStatus = ref<"idle" | "loading" | "ready" | "error">("idle");
const pointCount = ref(0);
const modelName = ref("bun_zipper.ply");
const pointSize = ref(2);
const autoRotate = ref(true);
const shaderMode = ref<"neon" | "height" | "normal" | "pulse">("height");
const selectedModel = ref("bun_zipper.ply");
const isWebGLSupported = ref(true);
const modelMenuRef = ref<HTMLElement | null>(null);
const shaderMenuRef = ref<HTMLElement | null>(null);
const isModelMenuOpen = ref(false);
const isShaderMenuOpen = ref(false);

let scene: THREE.Scene | null = null;
let camera: THREE.PerspectiveCamera | null = null;
let renderer: THREE.WebGLRenderer | null = null;
let controls: OrbitControls | null = null;
let pointCloud: THREE.Points<THREE.BufferGeometry, THREE.ShaderMaterial> | null = null;
let animationFrameId: number | null = null;
let resizeObserver: ResizeObserver | null = null;
const clock = new THREE.Clock();

const MODEL_URL_CANDIDATES = [
  "/bunny/reconstruction/bun_zipper.ply",
  "/bunny/reconstruction/bun_zipper_res2.ply",
  "/bunny/reconstruction/bun_zipper_res3.ply",
  "/bunny/reconstruction/bun_zipper_res4.ply",
];

const shaderModeOptions: Array<{ value: "neon" | "height" | "normal" | "pulse"; label: string }> = [
  { value: "neon", label: "Neon" },
  { value: "height", label: "Height" },
  { value: "normal", label: "Normal" },
  { value: "pulse", label: "Pulse" },
];
const modelOptions = [
  { value: "bun_zipper.ply", label: "Bunny Zipper (High)" },
  { value: "bun_zipper_res2.ply", label: "Bunny Res2" },
  { value: "bun_zipper_res3.ply", label: "Bunny Res3" },
  { value: "bun_zipper_res4.ply", label: "Bunny Res4" },
];
const shaderModeLabelMap: Record<"neon" | "height" | "normal" | "pulse", string> = {
  neon: "Neon",
  height: "Height",
  normal: "Normal",
  pulse: "Pulse",
};

onMounted(async () => {
  if (!canvasContainerRef.value || !isWebGLAvailable()) {
    isWebGLSupported.value = false;
    return;
  }

  document.addEventListener("visibilitychange", handleVisibilityChange);
  document.addEventListener("click", handleClickOutsideMenu);
  setupScene(canvasContainerRef.value);
  await loadPointCloud();
  startRenderLoop();
});

onUnmounted(() => {
  document.removeEventListener("visibilitychange", handleVisibilityChange);
  document.removeEventListener("click", handleClickOutsideMenu);
  disposeScene();
});

function handleBackToTasks() {
  router.push("/tasks");
}

function setupScene(container: HTMLElement) {
  // 基础场景与相机参数：保持轻量，不影响现有页面逻辑。
  scene = new THREE.Scene();
  scene.background = new THREE.Color("#050a0f");

  camera = new THREE.PerspectiveCamera(
    60,
    container.clientWidth / Math.max(container.clientHeight, 1),
    0.01,
    100,
  );
  camera.position.set(0.08, 0.12, 0.25);

  renderer = new THREE.WebGLRenderer({ antialias: true, alpha: false });
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
  renderer.setSize(container.clientWidth, container.clientHeight);
  container.appendChild(renderer.domElement);

  controls = new OrbitControls(camera, renderer.domElement);
  controls.enableDamping = true;
  controls.dampingFactor = 0.05;
  // 初始 target 保持在原点，后续在模型加载完成后按包围盒自动对齐。
  controls.target.set(0, 0, 0);

  // 环境光只用于辅助视觉，不参与点云着色结果。
  const ambientLight = new THREE.AmbientLight("#ffffff", 0.8);
  scene.add(ambientLight);

  resizeObserver = new ResizeObserver(() => {
    if (!canvasContainerRef.value || !camera || !renderer) {
      return;
    }
    camera.aspect =
      canvasContainerRef.value.clientWidth / Math.max(canvasContainerRef.value.clientHeight, 1);
    camera.updateProjectionMatrix();
    renderer.setSize(canvasContainerRef.value.clientWidth, canvasContainerRef.value.clientHeight);
  });
  resizeObserver.observe(container);
}

async function loadPointCloud() {
  if (!scene) {
    return;
  }

  clearCurrentPointCloud();
  renderStatus.value = "loading";
  const loader = new PLYLoader();
  const loaded = await tryLoadWithFallback(loader, selectedModel.value);
  if (!loaded) {
    renderStatus.value = "error";
    return;
  }

  try {
    const { geometry, source } = loaded;
    geometry.computeBoundingBox();
    geometry.center();

    // 使用统一 shader，通过 mode 在片元阶段切换配色策略。
    const material = new THREE.ShaderMaterial({
      uniforms: {
        uTime: { value: 0 },
        uPointSize: { value: pointSize.value },
        uColorA: { value: new THREE.Color("#00f3ff") },
        uColorB: { value: new THREE.Color("#39ff14") },
        uMode: { value: 1.0 },
      },
      vertexShader: `
        varying float vHeight;
        varying vec3 vNormal;
        uniform float uPointSize;

        void main() {
          vHeight = position.y;
          vNormal = normalize(normalMatrix * normal);
          vec4 mvPosition = modelViewMatrix * vec4(position, 1.0);
          gl_Position = projectionMatrix * mvPosition;
          gl_PointSize = uPointSize;
        }
      `,
      fragmentShader: `
        varying float vHeight;
        varying vec3 vNormal;
        uniform float uTime;
        uniform vec3 uColorA;
        uniform vec3 uColorB;
        uniform float uMode;

        void main() {
          float dist = distance(gl_PointCoord, vec2(0.5));
          if (dist > 0.5) {
            discard;
          }

          // mode:
          // 0.0 -> neon, 1.0 -> height, 2.0 -> normal, 3.0 -> pulse
          float mode = floor(uMode + 0.5);
          float heightT = clamp((vHeight + 0.08) / 0.16, 0.0, 1.0);
          float pulseT = 0.5 + 0.5 * sin(uTime * 2.2 + vHeight * 24.0);
          float normalT = clamp(vNormal.z * 0.5 + 0.5, 0.0, 1.0);
          vec3 color = mix(uColorA, uColorB, heightT);

          if (mode < 0.5) {
            color = uColorA;
          } else if (mode < 1.5) {
            color = mix(uColorA, uColorB, heightT);
          } else if (mode < 2.5) {
            color = mix(vec3(0.08, 0.18, 0.52), vec3(0.96, 0.98, 1.0), normalT);
          } else {
            color = mix(uColorA, uColorB, pulseT);
          }

          gl_FragColor = vec4(color, 1.0);
        }
      `,
      transparent: true,
      depthWrite: false,
    });

    pointCloud = new THREE.Points(geometry, material);
    scene.add(pointCloud);
    fitCameraToPointCloud(geometry);

    modelName.value = source.split("/").pop() ?? modelName.value;
    const positionAttr = geometry.getAttribute("position");
    pointCount.value = positionAttr ? positionAttr.count : 0;
    applyShaderPreset(shaderMode.value);
    renderStatus.value = "ready";
  } catch {
    renderStatus.value = "error";
  }
}

async function tryLoadWithFallback(loader: PLYLoader, preferredModel: string) {
  // 逐个尝试候选模型，避免因单个文件缺失导致整页不可用。
  const prioritized = [
    `/bunny/reconstruction/${preferredModel}`,
    ...MODEL_URL_CANDIDATES.filter((item) => item !== `/bunny/reconstruction/${preferredModel}`),
  ];
  for (const source of prioritized) {
    try {
      const geometry = await loader.loadAsync(source);
      return { geometry, source };
    } catch {
      // 继续尝试下一个文件。
    }
  }
  return null;
}

function clearCurrentPointCloud() {
  if (!pointCloud || !scene) {
    return;
  }
  pointCloud.geometry.dispose();
  pointCloud.material.dispose();
  scene.remove(pointCloud);
  pointCloud = null;
}

function fitCameraToPointCloud(geometry: THREE.BufferGeometry) {
  if (!camera || !controls) {
    return;
  }

  // 根据模型包围球计算最合适的观察距离，避免模型贴边或被裁切。
  geometry.computeBoundingSphere();
  const sphere = geometry.boundingSphere;
  if (!sphere) {
    return;
  }

  const center = sphere.center.clone();
  const radius = Math.max(sphere.radius, 0.001);
  const fov = THREE.MathUtils.degToRad(camera.fov);
  const fitDistance = radius / Math.tan(fov / 2);
  const distance = fitDistance * 1.35; // 留一点视觉边距

  camera.position.set(center.x, center.y, center.z + distance);
  camera.near = Math.max(0.001, distance / 100);
  camera.far = distance * 100;
  camera.updateProjectionMatrix();

  controls.target.copy(center);
  controls.minDistance = distance * 0.35;
  controls.maxDistance = distance * 4;
  controls.update();
}

function startRenderLoop() {
  const tick = () => {
    if (!scene || !camera || !renderer) {
      return;
    }

    animationFrameId = window.requestAnimationFrame(tick);
    controls?.update();

    if (pointCloud) {
      if (autoRotate.value) {
        pointCloud.rotation.y += 0.0025;
      }
      pointCloud.material.uniforms.uTime.value = clock.getElapsedTime();
      pointCloud.material.uniforms.uPointSize.value = pointSize.value;
    }

    renderer.render(scene, camera);
  };

  tick();
}

function handleModelSelect(nextModel: string) {
  selectedModel.value = nextModel;
  isModelMenuOpen.value = false;
  void loadPointCloud();
}

function handleShaderModeSelect(nextMode: "neon" | "height" | "normal" | "pulse") {
  shaderMode.value = nextMode;
  isShaderMenuOpen.value = false;
  applyShaderPreset(nextMode);
}

function toggleModelMenu() {
  isModelMenuOpen.value = !isModelMenuOpen.value;
  if (isModelMenuOpen.value) {
    isShaderMenuOpen.value = false;
  }
}

function toggleShaderMenu() {
  isShaderMenuOpen.value = !isShaderMenuOpen.value;
  if (isShaderMenuOpen.value) {
    isModelMenuOpen.value = false;
  }
}

function handleClickOutsideMenu(event: MouseEvent) {
  if (!(event.target instanceof Node)) {
    return;
  }
  if (modelMenuRef.value && !modelMenuRef.value.contains(event.target)) {
    isModelMenuOpen.value = false;
  }
  if (shaderMenuRef.value && !shaderMenuRef.value.contains(event.target)) {
    isShaderMenuOpen.value = false;
  }
}

function applyShaderPreset(mode: "neon" | "height" | "normal" | "pulse") {
  if (!pointCloud) {
    return;
  }

  const uniforms = pointCloud.material.uniforms;
  if (mode === "neon") {
    uniforms.uMode.value = 0.0;
    uniforms.uColorA.value = new THREE.Color("#00f3ff");
    uniforms.uColorB.value = new THREE.Color("#00f3ff");
    return;
  }
  if (mode === "height") {
    uniforms.uMode.value = 1.0;
    uniforms.uColorA.value = new THREE.Color("#00f3ff");
    uniforms.uColorB.value = new THREE.Color("#39ff14");
    return;
  }
  if (mode === "normal") {
    uniforms.uMode.value = 2.0;
    uniforms.uColorA.value = new THREE.Color("#00f3ff");
    uniforms.uColorB.value = new THREE.Color("#ffffff");
    return;
  }

  uniforms.uMode.value = 3.0;
  uniforms.uColorA.value = new THREE.Color("#00f3ff");
  uniforms.uColorB.value = new THREE.Color("#bc13fe");
}

function handleVisibilityChange() {
  // 页面切到后台时暂停渲染循环，避免无效 GPU 占用。
  if (document.hidden) {
    if (animationFrameId !== null) {
      window.cancelAnimationFrame(animationFrameId);
      animationFrameId = null;
    }
    return;
  }
  if (animationFrameId === null && renderer && camera && scene) {
    startRenderLoop();
  }
}

function isWebGLAvailable() {
  try {
    const canvas = document.createElement("canvas");
    return Boolean(canvas.getContext("webgl") || canvas.getContext("experimental-webgl"));
  } catch {
    return false;
  }
}

function disposeScene() {
  if (animationFrameId !== null) {
    window.cancelAnimationFrame(animationFrameId);
    animationFrameId = null;
  }

  resizeObserver?.disconnect();
  resizeObserver = null;

  clearCurrentPointCloud();

  controls?.dispose();
  controls = null;

  renderer?.dispose();
  if (renderer?.domElement.parentNode) {
    renderer.domElement.parentNode.removeChild(renderer.domElement);
  }
  renderer = null;
  camera = null;
  scene = null;
}
</script>

<template>
  <main class="min-h-screen bg-[#050a0f] text-white">
    <section class="relative mx-auto max-w-[1400px] px-4 py-6 md:px-6 md:py-8">
      <button
        class="absolute left-4 top-6 flex items-center gap-2 rounded-full border border-white/20 px-4 py-2 text-sm text-white/85 transition-colors hover:border-neon hover:text-neon md:left-6 md:top-8"
        @click="handleBackToTasks"
      >
        <ArrowLeft class="h-4 w-4" />
        {{ t("pointCloud.backToTasks") }}
      </button>

      <header class="mb-6 flex items-center justify-center">
        <h1 class="text-neon text-center text-3xl font-semibold tracking-[0.18em] drop-shadow-[0_0_18px_var(--neon-glow)] md:text-4xl">
          {{ t("pointCloud.title") }}
        </h1>
      </header>

      <div class="grid gap-4">
        <div class="flex flex-wrap items-center gap-4 rounded-xl border border-white/10 bg-white/5 px-4 py-3 text-sm text-white/80 md:flex-nowrap md:gap-5">
          <div
            ref="modelMenuRef"
            class="relative flex w-full shrink-0 items-center gap-2 md:w-[260px]"
          >
            <p class="shrink-0 whitespace-nowrap text-white/70">{{ t("pointCloud.modelLabel") }}:</p>
            <button
              class="flex min-w-0 flex-1 items-center justify-between rounded-full border border-white/10 px-2.5 py-1.5 transition-colors hover:border-white/25"
              @click="toggleModelMenu"
            >
              <span class="truncate text-sm text-white/85">
                {{ modelOptions.find((item) => item.value === selectedModel)?.label || selectedModel }}
              </span>
              <ChevronDown class="h-4 w-4 text-white/45" />
            </button>
            <div
              v-if="isModelMenuOpen"
              class="absolute left-[76px] top-10 z-20 w-[calc(100%-76px)] rounded-xl border border-white/10 bg-[#0b1219]/95 p-1.5 shadow-[0_14px_40px_rgba(0,0,0,0.35)]"
            >
              <button
                v-for="model in modelOptions"
                :key="model.value"
                class="w-full truncate rounded-lg px-3 py-2 text-left text-sm transition-colors"
                :class="selectedModel === model.value ? 'bg-white/10 text-white' : 'text-white/75 hover:bg-white/8'"
                :title="model.label"
                @click="handleModelSelect(model.value)"
              >
                {{ model.label }}
              </button>
            </div>
          </div>

          <p
            class="flex w-full shrink-0 items-center gap-1 whitespace-nowrap md:w-[180px]"
            :title="modelName"
          >
            <span>{{ t("pointCloud.datasetLabel") }}:</span>
            <span class="min-w-0 truncate text-neon">{{ modelName }}</span>
          </p>

          <div
            ref="shaderMenuRef"
            class="relative flex w-full shrink-0 items-center gap-2 md:w-[230px]"
          >
            <p class="shrink-0 whitespace-nowrap text-white/70">{{ t("pointCloud.shaderModeLabel") }}:</p>
            <button
              class="flex min-w-0 flex-1 items-center justify-between rounded-full border border-white/10 px-2.5 py-1.5 transition-colors hover:border-white/25"
              @click="toggleShaderMenu"
            >
              <span class="truncate text-sm text-white/85">{{ shaderModeLabelMap[shaderMode] }}</span>
              <ChevronDown class="h-4 w-4 text-white/45" />
            </button>
            <div
              v-if="isShaderMenuOpen"
              class="absolute left-[100px] top-10 z-20 w-[calc(100%-100px)] rounded-xl border border-white/10 bg-[#0b1219]/95 p-1.5 shadow-[0_14px_40px_rgba(0,0,0,0.35)]"
            >
              <button
                v-for="mode in shaderModeOptions"
                :key="mode.value"
                class="w-full rounded-lg px-3 py-2 text-left text-sm transition-colors"
                :class="shaderMode === mode.value ? 'bg-white/10 text-white' : 'text-white/75 hover:bg-white/8'"
                @click="handleShaderModeSelect(mode.value)"
              >
                {{ mode.label }}
              </button>
            </div>
          </div>

          <label class="flex w-full shrink-0 items-center gap-2 md:w-[210px]">
            <span class="whitespace-nowrap">{{ t("pointCloud.pointSizeLabel") }}: {{ pointSize.toFixed(1) }}</span>
            <input
              v-model.number="pointSize"
              type="range"
              min="1"
              max="5"
              step="0.1"
              class="w-28 accent-cyan-300"
            />
          </label>
          <label class="flex w-full shrink-0 items-center gap-2 md:w-[140px]">
            <input
              v-model="autoRotate"
              type="checkbox"
              class="h-4 w-4 rounded border-white/30 bg-[#0b1219]"
            />
            <span class="whitespace-nowrap">{{ t("pointCloud.autoRotateLabel") }}</span>
          </label>
        </div>

        <div class="relative h-[72vh] min-h-[440px] w-full rounded-2xl border border-white/10 bg-[#02070c]">
          <div
            class="absolute right-3 top-3 z-10 space-y-1 rounded-lg border border-white/10 bg-[#0b1219]/75 px-3 py-2 text-xs text-white/80 backdrop-blur-sm"
          >
            <p class="whitespace-nowrap">
              {{ t("pointCloud.statusLabel") }}:
              <span class="text-white">{{ t(`pointCloud.status.${renderStatus}`) }}</span>
            </p>
            <p class="whitespace-nowrap">
              {{ t("pointCloud.pointsLabel") }}:
              <span class="text-white">{{ pointCount }}</span>
            </p>
          </div>
          <div
            v-if="!isWebGLSupported"
            class="flex h-full items-center justify-center px-6 text-center text-white/70"
          >
            {{ t("pointCloud.webglUnsupported") }}
          </div>
          <div
            v-else
            ref="canvasContainerRef"
            class="h-full w-full"
          />
        </div>
      </div>
    </section>
  </main>
</template>
