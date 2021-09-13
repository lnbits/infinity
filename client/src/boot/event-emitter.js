import EventEmitter from 'events'

export default ({app}) => {
  app.config.globalProperties.$events = new EventEmitter()
}
