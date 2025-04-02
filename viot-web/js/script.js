// 等待DOM加載完成
document.addEventListener('DOMContentLoaded', function() {
  // 獲取DOM元素
  const pduCells = document.querySelectorAll('.pdu-cell');
  const bindDialog = document.getElementById('bindDialog');
  const unbindDialog = document.getElementById('unbindDialog');
  const manualBindDialog = document.getElementById('manualBindDialog');
  const confirmBindBtn = document.getElementById('confirmBindBtn');
  const cancelBindBtn = document.getElementById('cancelBindBtn');
  const confirmUnbindBtn = document.getElementById('confirmUnbindBtn');
  const cancelUnbindBtn = document.getElementById('cancelUnbindBtn');
  const confirmManualBindBtn = document.getElementById('confirmManualBindBtn');
  const cancelManualBindBtn = document.getElementById('cancelManualBindBtn');
  const manualBindButton = document.querySelector('.manual-bind-button');
  const checkboxes = document.querySelectorAll('.checkbox-container input');
  const tabButtons = document.querySelectorAll('.tab-button');
  const dcSegments = document.querySelectorAll('.segment');

  // 初始化數據
  let selectedPduIp = null;

  // DC下拉選單功能
  const dcDropdown = document.querySelector('.dc-dropdown');
  const dcSelected = document.querySelector('.dc-selected');
  const dcOptions = document.querySelector('.dc-options');
  const dcOptionItems = document.querySelectorAll('.dc-option');
  
  // 點擊選擇區域顯示/隱藏選項
  dcSelected.addEventListener('click', function(e) {
    e.stopPropagation();
    dcOptions.style.display = dcOptions.style.display === 'block' ? 'none' : 'block';
    
    // 切換箭頭方向
    const dropdownIcon = this.querySelector('.dropdown-icon');
    dropdownIcon.style.transform = dcOptions.style.display === 'block' ? 'rotate(180deg)' : 'rotate(0deg)';
  });
  
  // 點擊選項
  dcOptionItems.forEach(option => {
    option.addEventListener('click', function(e) {
      e.stopPropagation();
      
      // 更新選中顯示
      dcSelected.querySelector('span').textContent = this.textContent;
      
      // 移除其他選項的選中狀態
      dcOptionItems.forEach(opt => opt.classList.remove('selected'));
      
      // 添加當前選項的選中狀態
      this.classList.add('selected');
      
      // 隱藏選項列表
      dcOptions.style.display = 'none';
      
      // 重置箭頭方向
      const dropdownIcon = dcSelected.querySelector('.dropdown-icon');
      dropdownIcon.style.transform = 'rotate(0deg)';
      
      console.log('已選擇:', this.textContent);
    });
  });
  
  // 點擊頁面其他區域關閉下拉選單
  document.addEventListener('click', function() {
    dcOptions.style.display = 'none';
    
    // 重置箭頭方向
    const dropdownIcon = dcSelected.querySelector('.dropdown-icon');
    if (dropdownIcon) {
      dropdownIcon.style.transform = 'rotate(0deg)';
    }
  });

  // 為DC切換按鈕添加點擊事件
  dcSegments.forEach(segment => {
    segment.addEventListener('click', function() {
      dcSegments.forEach(s => s.classList.remove('selected'));
      this.classList.add('selected');
      // 這裡可以添加切換DC的邏輯
    });
  });

  // 為Room切換按鈕添加點擊事件
  tabButtons.forEach(button => {
    button.addEventListener('click', function() {
      tabButtons.forEach(b => b.classList.remove('active'));
      this.classList.add('active');
      // 這裡可以添加切換Room的邏輯
    });
  });

  // 為PDU複選框添加點擊事件
  checkboxes.forEach(checkbox => {
    checkbox.addEventListener('change', function() {
      // 單選邏輯
      if (this.checked) {
        checkboxes.forEach(cb => {
          if (cb !== this) cb.checked = false;
        });
        
        // 獲取選中的PDU IP
        const pduItem = this.closest('.pdu-item');
        selectedPduIp = pduItem.querySelector('.ip-text').textContent;
        
        // 更新詳細資訊卡
        updatePduDetailsCard(selectedPduIp);
        
        // 高亮顯示選中的PDU項目
        document.querySelectorAll('.pdu-item').forEach(item => {
          item.classList.remove('selected');
        });
        pduItem.classList.add('selected');
      } else {
        selectedPduIp = null;
        
        // 移除選中效果
        this.closest('.pdu-item').classList.remove('selected');
      }
    });
  });
  
  // 點擊PDU項目也可以選中/取消選中複選框
  document.querySelectorAll('.pdu-item').forEach(item => {
    item.addEventListener('click', function(e) {
      // 確保不是點擊在複選框上（已有事件處理）
      if (!e.target.closest('.checkbox-container')) {
        const checkbox = this.querySelector('input[type="checkbox"]');
        checkbox.checked = !checkbox.checked;
        
        // 手動觸發change事件
        const event = new Event('change');
        checkbox.dispatchEvent(event);
      }
    });
  });

  // PDU格子點擊事件
  const pduCellsNotEmpty = document.querySelectorAll('.pdu-cell:not(.empty)');
  pduCellsNotEmpty.forEach(cell => {
    cell.addEventListener('click', function() {
      const isMonitored = this.classList.contains('monitored');
      
      if (isMonitored) {
        // 已監控的PDU，顯示解除綁定對話框
        document.getElementById('unbindRack').textContent = this.textContent;
        toggleDialog('unbindDialog', true);
      } else {
        // 未監控的PDU，檢查是否有選中的PDU IP
        const checkedPdus = document.querySelectorAll('.pdu-item input[type="checkbox"]:checked');
        
        if (checkedPdus.length > 0) {
          // 顯示綁定對話框
          const pduIp = checkedPdus[0].closest('.pdu-item').querySelector('.ip-text').textContent;
          document.getElementById('bindIp').textContent = pduIp;
          document.getElementById('bindRack').textContent = this.textContent;
          toggleDialog('bindDialog', true);
        } else {
          alert('請先選擇左側的PDU IP');
        }
      }
    });
  });

  // 手動綁定按鈕
  const manualBindBtn = document.querySelector('.manual-bind-button');
  manualBindBtn.addEventListener('click', function() {
    toggleDialog('manualBindDialog', true);
  });

  // 對話框相關按鈕
  document.getElementById('cancelBindBtn').addEventListener('click', () => toggleDialog('bindDialog', false));
  document.getElementById('confirmBindBtn').addEventListener('click', confirmBind);
  
  document.getElementById('cancelUnbindBtn').addEventListener('click', () => toggleDialog('unbindDialog', false));
  document.getElementById('confirmUnbindBtn').addEventListener('click', confirmUnbind);
  
  document.getElementById('cancelManualBindBtn').addEventListener('click', () => toggleDialog('manualBindDialog', false));
  document.getElementById('confirmManualBindBtn').addEventListener('click', confirmManualBind);

  // 點擊對話框背景關閉對話框
  document.querySelectorAll('.dialog').forEach(dialog => {
    dialog.addEventListener('click', function(e) {
      if (e.target === this) hideDialog(this);
    });
  });

  // 顯示/隱藏對話框
  function toggleDialog(dialogId, show) {
    const dialog = document.getElementById(dialogId);
    if (show) {
      dialog.classList.add('show');
    } else {
      dialog.classList.remove('show');
    }
  }

  // 確認綁定
  function confirmBind() {
    const pduIp = document.getElementById('bindIp').textContent;
    const rackName = document.getElementById('bindRack').textContent;
    
    console.log(`綁定PDU: ${pduIp} 到 ${rackName}`);
    
    // 模擬綁定成功
    const pduCells = document.querySelectorAll('.pdu-cell');
    pduCells.forEach(cell => {
      if (cell.textContent === rackName) {
        cell.classList.add('monitored');
      }
    });
    
    // 關閉對話框
    toggleDialog('bindDialog', false);
  }

  // 確認解除綁定
  function confirmUnbind() {
    const rackName = document.getElementById('unbindRack').textContent;
    
    console.log(`解除綁定: ${rackName}`);
    
    // 模擬解除綁定
    const pduCells = document.querySelectorAll('.pdu-cell');
    pduCells.forEach(cell => {
      if (cell.textContent === rackName) {
        cell.classList.remove('monitored');
      }
    });
    
    // 關閉對話框
    toggleDialog('unbindDialog', false);
  }

  // 確認手動綁定
  function confirmManualBind() {
    const rackName = document.getElementById('rackName').value;
    const side = document.querySelector('input[name="side"]:checked').value;
    const fullRackName = `${rackName} ${side}`;
    
    console.log(`手動綁定到: ${fullRackName}`);
    
    // 檢查是否存在該機櫃
    let found = false;
    const pduCells = document.querySelectorAll('.pdu-cell');
    pduCells.forEach(cell => {
      if (cell.textContent === fullRackName) {
        cell.classList.add('monitored');
        found = true;
      }
    });
    
    if (!found) {
      alert(`找不到機櫃: ${fullRackName}`);
    }
    
    // 關閉對話框
    toggleDialog('manualBindDialog', false);
  }

  // 更新PDU詳細資訊卡
  function updatePduDetailsCard(ip) {
    // 模擬從後端獲取PDU詳細資訊
    // 這裡使用靜態數據
    const pduDetails = {
      ip: ip,
      room: 'R5',
      mfg: 'Deata',
      model: 'PDUE428',
      sn: 'Y3456789'
    };
    
    // 更新詳細資訊卡
    const detailItems = document.querySelectorAll('.detail-item .value');
    detailItems[0].textContent = pduDetails.ip;
    detailItems[1].textContent = pduDetails.room;
    detailItems[2].textContent = pduDetails.mfg;
    detailItems[3].textContent = pduDetails.model;
    detailItems[4].textContent = pduDetails.sn;
  }

  // 模擬自動刷新PDU狀態
  function autoRefreshPduStatus() {
    console.log('自動刷新PDU狀態...');
    // 這裡可以添加定期從後端獲取PDU狀態的邏輯
  }

  // 設置每10秒刷新一次PDU狀態
  setInterval(autoRefreshPduStatus, 10000);
}); 