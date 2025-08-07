<script setup>
import { onMounted, ref } from 'vue'
import HelloWorld from './components/HelloWorld.vue'
import SearchToolbar from './components/SearchToolbar.vue'
import ComicCard from './components/ComicCard.vue'
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




</script>

<template>
  <div id="app" class="drop-target">
   

    <div class="content">
      <div v-if="loading" class="loading">加载中...</div>
      
      <div v-else-if="comics.length === 0" class="empty-state">
        <p>还没有漫画，拖放文件夹或zip文件到这里开始添加！</p>
      </div>
      
      <div v-else class="comics-grid">
        <ComicCard 
          v-for="comic in comics" 
          :key="comic.id" 
          :comic="comic"
          @delete="deleteComic"
        />
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
