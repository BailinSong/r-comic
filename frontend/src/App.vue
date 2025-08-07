<script setup>
import { onMounted, ref } from 'vue'
import HelloWorld from './components/HelloWorld.vue'
import SearchToolbar from './components/SearchToolbar.vue'
import { initFileDrop } from './dragAndDrop.js'
import { HandleFileDrop, GetComicsFromDatabase, SearchComicsInDatabase, DeleteComicFromDatabase, GetImageBase64 } from '../wailsjs/go/main/App.js'


const comics = ref([])
const searchKeyword = ref('')
const loading = ref(false)


onMounted(() => {
  // 初始化文件拖放功能
  initFileDrop(HandleFileDrop)
  // 加载漫画列表
  loadComics()
  

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

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}


</script>

<template>
  <div id="app" class="drop-target">
   

    <div class="content">
      <div v-if="loading" class="loading">加载中...</div>
      
      <div v-else-if="comics.length === 0" class="empty-state">
        <p>还没有漫画，拖放文件夹或zip文件到这里开始添加！</p>
      </div>
      
      <div v-else class="comics-grid">
        <div 
          v-for="comic in comics" 
          :key="comic.id" 
          class="comic-card"
        >
          <!-- 图片预览区域 -->
          <div class="comic-preview">
            <img :src="`${comic.firstImage}`" :alt="comic.title" class="comic-image"/>
            <!-- <img 
              v-if="comic.firstImage && imageCache.get(comic.firstImage)"
              :src="comic.firstImage"
              :alt="comic.title"
              class="comic-image"
              @error="handleImageError"
            />
            <div v-else-if="comic.firstImage && imageLoading.has(comic.firstImage)" class="loading-image">
              <span>加载中...</span>
            </div>
            <div v-else-if="comic.firstImage" class="no-image">
              <span>图片加载失败</span>
            </div>
            <div v-else class="no-image">
              <span>暂无图片</span>
            </div> -->
          </div>
          
          <div class="comic-info">
            <h3>{{ comic.title }}</h3>
            <p><strong>类型:</strong> {{ comic.fileType }}</p>
            <p><strong>大小:</strong> {{ formatFileSize(comic.fileSize) }}</p>
            <p><strong>第一张图片:</strong> {{ comic.firstImage }}</p>
            <p><strong>路径:</strong> {{ comic.filePath }}</p>
            <p><strong>添加时间:</strong> {{ new Date(comic.createdAt).toLocaleString() }}</p>
          </div>
          <div class="comic-actions">
            <button @click="deleteComic(comic.id)" class="delete-btn">删除</button>
          </div>
        </div>
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
  padding: 20px;
  padding-bottom: 100px; /* 为悬浮工具栏留出空间 */
  max-width: 1200px;
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

/* 漫画网格 */
.comics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
}

/* 漫画卡片 */
.comic-card {
  background: white;
  border-radius: 10px;
  padding: 20px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
  transition: transform 0.2s, box-shadow 0.2s;
  display: flex;
  flex-direction: column;
}

.comic-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0,0,0,0.15);
}

/* 图片预览区域 */
.comic-preview {
  margin-bottom: 15px;
  border-radius: 8px;
  overflow: hidden;
  background: #f8f9fa;
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.comic-image {
  width: 100%;
  height: 200px;
  object-fit: cover;
  border-radius: 8px;
  transition: transform 0.3s;
}

.comic-image:hover {
  transform: scale(1.05);
}

.no-image, .loading-image {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: #999;
  font-size: 14px;
  background: #f8f9fa;
  border: 2px dashed #ddd;
  border-radius: 8px;
}

.loading-image {
  color: #007bff;
  border-color: #007bff;
  border-style: solid;
}

.comic-info h3 {
  color: #333;
  margin-bottom: 15px;
  font-size: 1.2rem;
  font-weight: 600;
}

.comic-info p {
  margin-bottom: 8px;
  color: #666;
  font-size: 14px;
  line-height: 1.4;
}

.comic-info strong {
  color: #333;
}

.comic-actions {
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #eee;
}

.delete-btn {
  background: #dc3545;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
  transition: background 0.3s;
}

.delete-btn:hover {
  background: #c82333;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .header h1 {
    font-size: 1.5rem;
  }
  
  .search-box {
    flex-direction: column;
  }
  
  .comics-grid {
    grid-template-columns: 1fr;
  }
}
</style>
