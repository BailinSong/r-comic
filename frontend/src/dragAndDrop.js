import { OnFileDrop } from '../wailsjs/runtime/runtime.js';

// 初始化文件拖放功能
export function initFileDrop(handleFileDrop) {
    OnFileDrop((x, y, paths) => {
        console.log(`文件拖放到位置: (${x}, ${y})`);
        console.log('文件路径:', paths);
        
        // 调用 Go 后端的方法处理文件
        if (handleFileDrop && typeof handleFileDrop === 'function') {
            handleFileDrop(paths);
        }
    }, false); // 使用 drop target 样式
} 