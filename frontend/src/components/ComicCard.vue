<template>
  <div class="comic-card">
    <!-- 图片预览区域 -->
    <div class="comic-preview">
      <img :src="`${comic.firstImage}`" :alt="comic.title" class="comic-image"/>
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
      <button @click="handleDelete" class="delete-btn">删除</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ComicCard',
  props: {
    comic: {
      type: Object,
      required: true
    }
  },
  emits: ['delete'],
  methods: {
    formatFileSize(bytes) {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    },
    handleDelete() {
      this.$emit('delete', this.comic.id)
    }
  }
}
</script>

<style scoped>
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
</style>
