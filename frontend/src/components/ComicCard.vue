<template>
  <div class="comic-card" @click="handleOpen" :title="comic.title">
    <div class="cover">
      <img
        v-if="imgSrc && !hasError"
        :src="imgSrc"
        :alt="comic.title"
        class="cover-img"
        @load="onImageLoaded"
        @error="onImageError"
      />

      <div v-else class="cover-fallback">
        <svg viewBox="0 0 24 24" class="fallback-icon" aria-hidden="true">
          <path fill="currentColor" d="M19 3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V5a2 2 0 0 0-2-2m0 16H5V5h14zM7 15l2.5-3.15L12 15l3.5-4.5L19 15z"/>
        </svg>
        <span class="fallback-text">无预览</span>
      </div>

      <!-- 骨架屏 -->
      <div v-if="!imageLoaded && imgSrc && !hasError" class="skeleton" />

      <!-- 顶部信息徽标 -->
      <div class="badges">
        <span class="badge type" :data-type="comic.fileType">{{ typeLabel }}</span>
        <span v-if="comic.fileSize" class="badge size">{{ formattedSize }}</span>
      </div>

      <!-- 底部标题渐变层 -->
      <div class="title-bar">
        <h3 class="title" :title="comic.title">{{ comic.title }}</h3>
      </div>

      <!-- 悬浮操作条 -->
      <div class="hover-actions" @click.stop>
        <button class="btn primary" @click="handleOpen">
          <svg viewBox="0 0 24 24" class="btn-icon" aria-hidden="true">
            <path fill="currentColor" d="M13 9h5.5L12 3.5 5.5 9H11v6h2zM5 18h14v2H5z"/>
          </svg>
          查看
        </button>
        <button class="btn danger" @click="handleDelete">
          <svg viewBox="0 0 24 24" class="btn-icon" aria-hidden="true">
            <path fill="currentColor" d="M9 3v1H4v2h16V4h-5V3H9m1 6v8h2V9h-2m-4 0v8h2V9H6m8 0v8h2V9h-2z"/>
          </svg>
          删除
        </button>
      </div>
    </div>

    <!-- 次要信息 -->
    <div class="meta">
      <div class="meta-item" :title="comic.filePath">
        <svg viewBox="0 0 24 24" class="meta-icon" aria-hidden="true">
          <path fill="currentColor" d="M10 4H4v16h16V8h-8zM5 19V5h4v4h10v10z"/>
        </svg>
        <span class="path">{{ shortPath }}</span>
      </div>
      <div class="meta-item" v-if="comic.createdAt">
        <svg viewBox="0 0 24 24" class="meta-icon" aria-hidden="true">
          <path fill="currentColor" d="M19 4h-1V2h-2v2H8V2H6v2H5a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2m0 15H5V10h14z"/>
        </svg>
        <span>{{ createdAtLabel }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'

const props = defineProps({
  comic: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['delete', 'open'])

const hasError = ref(false)
const imageLoaded = ref(false)

const imgSrc = computed(() => {
  const src = props.comic?.firstImage || ''
  return typeof src === 'string' ? src : ''
})

const formattedSize = computed(() => formatFileSize(props.comic?.fileSize || 0))
const typeLabel = computed(() => (props.comic?.fileType || '').toUpperCase())
const createdAtLabel = computed(() => {
  const v = props.comic?.createdAt
  try {
    return v ? new Date(v).toLocaleString() : ''
  } catch {
    return ''
  }
})

const shortPath = computed(() => {
  const p = props.comic?.filePath || ''
  if (p.length <= 42) return p
  return p.slice(0, 18) + '…' + p.slice(-21)
})

function formatFileSize(bytes) {
  if (!bytes || bytes <= 0) return '—'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(k)), sizes.length - 1)
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function handleDelete() {
  emit('delete', props.comic?.id)
}

function handleOpen() {
  emit('open', props.comic)
}

function onImageLoaded() {
  imageLoaded.value = true
}

function onImageError() {
  hasError.value = true
}
</script>

<style scoped>
.comic-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: #fff;
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 2px 10px rgba(0,0,0,0.06);
  transition: transform .2s ease, box-shadow .2s ease;
  cursor: pointer;
  /* Masonry 下让卡片在列中自然撑开 */
  width: 100%;
}
.comic-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 10px 24px rgba(0,0,0,0.12);
}

.cover {
  position: relative;
  background: linear-gradient(180deg, #f6f7f9 0%, #eef1f5 100%);
  /* Masonry 下让高度由内容决定，避免固定纵横比导致列只有 1 个 */
  aspect-ratio: .618; 
  width: 100%;
}
.cover-img {
  display: block;
  inset: 0;                                                                                                                                          
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform .35s ease;
}
.comic-card:hover .cover-img {
  transform: scale(1.0);
}

.cover-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #8a8f98;
  gap: 10px;
  font-size: 14px;
  min-height: 200px;
}
.fallback-icon {
  width: 42px;
  height: 42px;
}
.fallback-text { opacity: .9; }

.skeleton {
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, #f2f3f5 25%, #e9ebef 37%, #f2f3f5 63%);
  background-size: 400% 100%;
  animation: shimmer 1.2s infinite;
}
@keyframes shimmer {
  0% { background-position: 100% 0; }
  100% { background-position: -100% 0; }
}

.badges {
  position: absolute;
  top: 10px;
  left: 10px;
  right: 10px;
  display: flex;
  justify-content: space-between;
  gap: 8px;
  pointer-events: none;
}
.badge {
  pointer-events: auto;
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  color: #1f2937;
  background: rgba(255,255,255,0.82);
  backdrop-filter: blur(6px);
  box-shadow: 0 4px 10px rgba(0,0,0,0.08);
}
.badge.type[data-type="zip"] { color: #6f2cff; }
.badge.type[data-type="folder"] { color: #0ea5e9; }
.badge.size { color: #111827; }

.title-bar {
  position: absolute;
  left: 0; right: 0; bottom: 0;
  padding: 14px 14px 12px;
  background: linear-gradient(180deg, rgba(0,0,0,0) 0%, rgba(0,0,0,0.55) 100%);
  color: #fff;
}
.title {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  text-shadow: 0 1px 10px rgba(0,0,0,0.35);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hover-actions {
  position: absolute;
  inset: auto 12px 12px auto;
  display: flex;
  gap: 10px;
  opacity: 0;
  transform: translateY(6px);
  transition: all .2s ease;
}
.comic-card:hover .hover-actions {
  opacity: 1;
  transform: translateY(0);
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 12px;
  border-radius: 8px;
  border: none;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(0,0,0,0.12);
}
.btn .btn-icon { width: 16px; height: 16px; }
.btn.primary { background: #4f46e5; color: #fff; }
.btn.primary:hover { background: #4338ca; }
.btn.danger { background: #ef4444; color: #fff; }
.btn.danger:hover { background: #dc2626; }

.meta {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 8px 10px;
  align-items: center;
  padding: 0 10px 12px;
  color: #6b7280;
}
.meta-item { display: inline-flex; align-items: center; gap: 6px; min-width: 0; }
.meta-icon { width: 16px; height: 16px; opacity: .8; }
.path { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

@media (max-width: 768px) {
  .title { font-size: 15px; }
  .btn { height: 30px; padding: 0 10px; }
}
</style>
