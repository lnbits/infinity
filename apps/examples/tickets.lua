function init ()
  return {
    models = {
      {
        name = 'Ticket Bucket',
        fields = {
          { name = 'price', type = 'number', required = true },
          { name = 'description', type = 'string' }
        }
      },
      {
        name = 'Ticket',
        fields = {
          { name = 'bucket', ref = 'Ticket Bucket' },
          { name = 'content', type = 'string', required = true },
          { name = 'author', type = 'string' }
        },
      },
    }
  }
end
