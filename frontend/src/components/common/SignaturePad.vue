<template>
  <div class="space-y-2">
    <div class="flex items-center justify-between">
      <label class="text-xs font-semibold text-slate-700">{{ label || 'Tanda Tangan Digital' }}</label>
      <button
        type="button"
        @click="clear"
        class="text-xs text-rose-600 hover:text-rose-700 font-medium flex items-center space-x-1"
      >
        <span>Bersihkan Tanda Tangan</span>
      </button>
    </div>

    <div class="border-2 border-dashed border-slate-300 rounded-2xl p-2 bg-slate-50 relative overflow-hidden">
      <canvas
        ref="canvasRef"
        class="w-full h-36 bg-white rounded-xl shadow-xs cursor-crosshair touch-none"
        @mousedown="startDrawing"
        @mousemove="draw"
        @mouseup="stopDrawing"
        @mouseleave="stopDrawing"
        @touchstart="handleTouchStart"
        @touchmove="handleTouchMove"
        @touchend="stopDrawing"
      ></canvas>

      <div
        v-if="isEmpty && !isDrawing"
        class="absolute inset-0 flex flex-col items-center justify-center pointer-events-none text-slate-400 text-xs select-none"
      >
        <svg class="w-6 h-6 mb-1 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
        </svg>
        <span>Goreskan tanda tangan di sini (Layar Sentuh / Mouse)</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const props = defineProps<{
  label?: string
}>()

const emit = defineEmits<{
  (e: 'update:signature', data: string): void
}>()

const canvasRef = ref<HTMLCanvasElement | null>(null)
const isDrawing = ref(false)
const isEmpty = ref(true)
let ctx: CanvasRenderingContext2D | null = null

function initCanvas() {
  const canvas = canvasRef.value
  if (!canvas) return
  ctx = canvas.getContext('2d')
  if (!ctx) return

  // Set high resolution
  const rect = canvas.getBoundingClientRect()
  canvas.width = rect.width * 2
  canvas.height = rect.height * 2
  ctx.scale(2, 2)

  ctx.lineWidth = 2.5
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.strokeStyle = '#0F172A'
}

function getCoords(e: MouseEvent) {
  const canvas = canvasRef.value
  if (!canvas) return { x: 0, y: 0 }
  const rect = canvas.getBoundingClientRect()
  return {
    x: e.clientX - rect.left,
    y: e.clientY - rect.top,
  }
}

function startDrawing(e: MouseEvent) {
  if (!ctx) return
  isDrawing.value = true
  isEmpty.value = false
  const { x, y } = getCoords(e)
  ctx.beginPath()
  ctx.moveTo(x, y)
}

function draw(e: MouseEvent) {
  if (!isDrawing.value || !ctx) return
  const { x, y } = getCoords(e)
  ctx.lineTo(x, y)
  ctx.stroke()
}

function stopDrawing() {
  if (!isDrawing.value) return
  isDrawing.value = false
  if (canvasRef.value) {
    const dataUrl = canvasRef.value.toDataURL('image/png')
    emit('update:signature', dataUrl)
  }
}

function handleTouchStart(e: TouchEvent) {
  if (e.touches.length === 0 || !ctx) return
  e.preventDefault()
  isDrawing.value = true
  isEmpty.value = false
  const touch = e.touches[0]
  const rect = canvasRef.value!.getBoundingClientRect()
  const x = touch.clientX - rect.left
  const y = touch.clientY - rect.top
  ctx.beginPath()
  ctx.moveTo(x, y)
}

function handleTouchMove(e: TouchEvent) {
  if (!isDrawing.value || !ctx || e.touches.length === 0) return
  e.preventDefault()
  const touch = e.touches[0]
  const rect = canvasRef.value!.getBoundingClientRect()
  const x = touch.clientX - rect.left
  const y = touch.clientY - rect.top
  ctx.lineTo(x, y)
  ctx.stroke()
}

function clear() {
  const canvas = canvasRef.value
  if (!canvas || !ctx) return
  ctx.clearRect(0, 0, canvas.width, canvas.height)
  isEmpty.value = true
  emit('update:signature', '')
}

function getSignature(): string {
  if (isEmpty.value || !canvasRef.value) return ''
  return canvasRef.value.toDataURL('image/png')
}

defineExpose({
  clear,
  getSignature,
  isEmpty,
})

onMounted(() => {
  initCanvas()
})
</script>
