<script setup>
import { onMounted, onBeforeUnmount, ref, computed, watch, nextTick } from 'vue'
import HelloWorld from './components/HelloWorld.vue'
import SearchToolbar from './components/SearchToolbar.vue'
import ComicCard from './components/ComicCard.vue'
import { initFileDrop } from './dragAndDrop.js'
import { HandleFileDrop, GetComicsFromDatabase, SearchComicsInDatabase, DeleteComicFromDatabase, GetImageBase64 } from '../wailsjs/go/main/App.js'


const comics = ref([])
const searchKeyword = ref('')
const loading = ref(false)

// 虚拟列表：增量渲染（批量加载）
const BATCH_SIZE = 40
const visibleCount = ref(BATCH_SIZE)
const visibleComics = computed(() => comics.value.slice(0, visibleCount.value))
const sentinel = ref(null)
const masonry = ref(null)
const CARD_MIN_WIDTH = 260
const GRID_GAP = 18
const GRID_ROW = 8

const masonryStyle = computed(() => ({
  '--col-min': `${CARD_MIN_WIDTH}px`,
  '--gap': `${GRID_GAP}px`,
  '--row': `${GRID_ROW}px`
}))
let io = null
let ro = null


onMounted(() => {
  // 初始化文件拖放功能
  initFileDrop(HandleFileDrop)
  // 加载漫画列表
  loadComics()
  // 初始化观察器：滚动接近底部时增加渲染数量
  setupIntersectionObserver()
  setupResizeObserver()
  setupImageLoadListeners()
  nextTick(layoutMasonry)
})

onBeforeUnmount(() => {
  if (io) {
    io.disconnect()
    io = null
  }
  if (ro) {
    ro.disconnect()
    ro = null
  }
})

const loadComics = async () => {
  console.log('[loadComics] 开始加载漫画列表')
  try {
    loading.value = true
    console.log('[loadComics] 从数据库获取漫画数据...')
    comics.value = await GetComicsFromDatabase()
    console.log(`[loadComics] 成功加载 ${comics.value.length} 本漫画`)

  } catch (error) {
    console.error('[loadComics] 加载漫画失败:', error)
    console.error('[loadComics] 错误详情:', {
      message: error.message,
      stack: error.stack
    })
  } finally {
    loading.value = false
    console.log('[loadComics] 加载状态已重置')
  }
}

watch(comics, () => {
  visibleCount.value = Math.min(BATCH_SIZE, comics.value.length)
})

function setupIntersectionObserver() {
  if (!('IntersectionObserver' in window)) return
  io = new IntersectionObserver((entries) => {
    const entry = entries[0]
    if (entry && entry.isIntersecting) {
      if (visibleCount.value < comics.value.length) {
        visibleCount.value = Math.min(visibleCount.value + BATCH_SIZE, comics.value.length)
      }
    }
  }, { root: null, rootMargin: '600px 0px', threshold: 0 })

  // 下一桢再观察，确保元素已渲染
  requestAnimationFrame(() => {
    if (sentinel.value && io) io.observe(sentinel.value)
  })
}

function setupResizeObserver() {
  const relayout = () => requestAnimationFrame(layoutMasonry)
  const attach = () => {
    if (!masonry.value) return
    if (ro) ro.disconnect()
    if ('ResizeObserver' in window) {
      ro = new ResizeObserver(relayout)
      ro.observe(masonry.value)
    } else {
      window.addEventListener('resize', relayout)
    }
    relayout()
  }
  requestAnimationFrame(attach)
  watch(masonry, () => requestAnimationFrame(attach))
  watch(visibleComics, () => requestAnimationFrame(layoutMasonry))
}

function setupImageLoadListeners() {
  const bind = () => {
    if (!masonry.value) return
    const imgs = masonry.value.querySelectorAll('img')
    imgs.forEach(img => {
      img.removeEventListener?.('load', layoutMasonry)
      img.addEventListener('load', layoutMasonry, { once: false })
    })
  }
  nextTick(bind)
  watch(visibleComics, () => nextTick(bind))
}

function layoutMasonry() {
  const container = masonry.value
  if (!container) return
  const gap = GRID_GAP
  const row = GRID_ROW
  const items = container.querySelectorAll('.masonry-item')
  items.forEach(el => {
    const height = el.offsetHeight
    const span = Math.max(1, Math.ceil((height + gap) / (row + gap)))
    el.style.gridRowEnd = `span ${span}`
  })
}

const searchComics = async () => {
  if (!searchKeyword.value.trim()) {
    await loadComics()
    return
  }
  
  try {
    loading.value = true
    comics.value = await SearchComicsInDatabase(searchKeyword.value)
  } catch (error) {
    console.error('搜索漫画失败:', error)
  } finally {
    loading.value = false
  }
}

const deleteComic = async (comicId) => {
  try {
    await DeleteComicFromDatabase(comicId)
    await loadComics() // 重新加载列表
  } catch (error) {
    console.error('删除漫画失败:', error)
  }
}




</script>

<template>
  <div id="app" class="drop-target">
   

    <div class="content">
      <div v-if="loading" class="loading">加载中...</div>
      
      <div v-else-if="comics.length === 0" class="empty-state">
        <p>还没有漫画，拖放文件夹或zip文件到这里开始添加！</p>
      </div>
      
      <div v-else class="masonry" ref="masonry" :style="masonryStyle">
        <div 
          v-for="comic in visibleComics" 
          :key="comic.id" 
          class="masonry-item"
          :style="{ gridRowEnd: 'span 1' }"
        >
          <ComicCard 
            :comic="comic"
            @delete="deleteComic"
          />
        </div>
        <!-- 加载更多的侦测器 -->
        <div ref="sentinel" class="masonry-sentinel"></div>
      </div>
    </div>

    <!-- 悬浮搜索工具栏 -->
    <SearchToolbar 
      @search="searchComics" 
      @refresh="loadComics"
    />
  </div>
</template>

<style>
/* 基础样式 */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background-color: #f5f5f5;
}

/* 拖放目标样式 */
.drop-target {
  --wails-drop-target: drop;
  min-height: 100vh;
  width: 100%;
  background-color: #f5f5f5;
  display: flex;
  flex-direction: column;
}

/* 拖放时的视觉反馈 */
.drop-target:has([dragover]) {
  background-color: rgba(0, 123, 255, 0.1);
  border: 2px dashed #007bff;
}





/* 头部样式 */
.header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 20px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}

.header h1 {
  margin-bottom: 20px;
  font-size: 2rem;
  font-weight: 300;
}

/* 内容区域 */
.content {
  padding: 24px;
  padding-bottom: 100px; /* 为悬浮工具栏留出空间 */
  max-width: 1440px;
  margin: 0 auto;
}

.loading {
  text-align: center;
  padding: 40px;
  font-size: 18px;
  color: #666;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #666;
}

.empty-state p {
  font-size: 18px;
  margin-bottom: 20px;
}

/* Masonry 布局：使用 CSS Grid 实现稳定的多列瀑布流 */
.masonry {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(var(--col-min, 260px), 1fr));
  grid-auto-rows: var(--row, 8px);
  gap: var(--gap, 18px);
}
.masonry-item {
  /* gridRowEnd 在运行时根据内容高度计算并设置 */
}
.masonry-sentinel {
  width: 100%;
  height: 1px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .header h1 {
    font-size: 1.5rem;
  }
  
  .search-box {
    flex-direction: column;
  }
  
  .masonry { column-count: 1; column-gap: 14px; }
}

@media (min-width: 1200px) {
  .content {
    max-width: 1520px;
  }
  .masonry { column-count: 6; column-gap: 20px; }
}

@media (min-width: 1600px) {
  .content {
    max-width: 1680px;
  }
  .masonry { column-count: 7; column-gap: 22px; }
}
</style>
