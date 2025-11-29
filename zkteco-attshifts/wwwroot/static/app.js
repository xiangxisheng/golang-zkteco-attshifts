document.addEventListener('DOMContentLoaded',()=>{
  const form=document.getElementById('ym-form')
  const year=form?.querySelector('select[name="year"]')
  const month=form?.querySelector('select[name="month"]')
  ;[year,month].forEach(el=>el&&el.addEventListener('change',()=>form.submit()))

  const loading=document.getElementById('loading')

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
    const fmt=dlForm.querySelector('select[name="fmt"]').value
    dlForm.action=fmt==='xls'?'/download.xls':'/download'
    modal.classList.add('hidden')
    loading?.classList.remove('hidden')
  })
})
