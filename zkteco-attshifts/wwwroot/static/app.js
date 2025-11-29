document.addEventListener('DOMContentLoaded',()=>{
  const form=document.getElementById('ym-form')
  const year=form?.querySelector('select[name="year"]')
  const month=form?.querySelector('select[name="month"]')
  ;[year,month].forEach(el=>el&&el.addEventListener('change',()=>form.submit()))

  const loading=document.getElementById('loading')
  const colsForm=document.getElementById('cols-form')
  const openCols=document.getElementById('open-cols')
  const colsModal=document.getElementById('cols-modal')
  const closeCols=document.getElementById('close-cols')
  openCols&&openCols.addEventListener('click',()=>{ colsModal.classList.remove('hidden') })
  closeCols&&closeCols.addEventListener('click',()=>{ colsModal.classList.add('hidden') })
  colsModal&&colsModal.addEventListener('click',e=>{ if(e.target===colsModal){ colsModal.classList.add('hidden') } })
  colsForm&&colsForm.addEventListener('submit',()=>{ loading?.classList.remove('hidden') })

  const dlForm=document.getElementById('dl-form')
  const openBtn=document.getElementById('open-dl')
  const modal=document.getElementById('dl-modal')
  const closeBtn=document.getElementById('close-dl')

  openBtn&&openBtn.addEventListener('click',()=>{
    modal.classList.remove('hidden')
  })
  closeBtn&&closeBtn.addEventListener('click',()=>{
    modal.classList.add('hidden')
  })
  modal&&modal.addEventListener('click',e=>{
    if(e.target===modal){ modal.classList.add('hidden') }
  })

  dlForm&&dlForm.addEventListener('submit',e=>{
    const fmt=(dlForm.querySelector('input[name="fmt"]:checked')||{value:'csv'}).value
    dlForm.action = (fmt==='xls') ? '/download.xls' : (fmt==='html' ? '/download.html' : '/download')
    modal.classList.add('hidden')
    loading?.classList.remove('hidden')
    setTimeout(()=>{ loading?.classList.add('hidden') }, 2000)
  })

  document.addEventListener('visibilitychange',()=>{
    if(document.visibilityState==='visible'){ loading?.classList.add('hidden') }
  })
})
