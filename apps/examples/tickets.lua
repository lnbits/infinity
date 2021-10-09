models = {
  {
    name = 'bucket',
    display = 'Ticket Bucket',
    fields = {
      { name = 'price', display = 'Price (msat)', type = 'number', required = true },
      { name = 'description', display = 'Description', type = 'string' },
      { display = 'URL', type = 'url', computed = function (bucket)
        return '/' .. bucket.key
      end }
    }
  },
  {
    name = 'ticket',
    display = 'Ticket',
    fields = {
      { name = 'bucket', display = 'Bucket', ref = 'Ticket Bucket' },
      { name = 'content', display = 'Content', type = 'string', required = true },
      { name = 'author', display = 'Author', type = 'string' },
      { name = 'is_paid', type = 'boolean', hidden = true }
    },
    filter = function (ticket)
      return not ticket.is_paid
    end
  },
}

actions = {
  create_ticket = function (params)
    s.create_invoice()

    models.ticket.add({
      bucket = params.bucket,

    })
  end
}

on = {
  payment_received = function (payment)

  end
}
