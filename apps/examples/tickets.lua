models = {
  {
    name = 'bucket',
    display = 'Ticket Bucket',
    fields = {
      { name = 'price', display = 'Price (msat)', type = 'msatoshi', required = true },
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
      { name = 'bucket', display = 'Bucket', type = 'ref', ref = 'bucket' },
      { name = 'content', display = 'Content', type = 'string', required = true },
      { name = 'author', display = 'Author', type = 'string' },
      { name = 'is_paid', type = 'boolean', hidden = true },
    },
    filter = function (ticket)
      return not ticket.is_paid
    end
  },
}

actions = {
  create_ticket = function (params)
    local key = db.ticket.add({
      bucket = params.bucket,
      content = params.content,
      author = params.author,
      is_paid = false,
    })

    local payment = s.create_invoice({
      extra = { ticket = key }
    })
  end
}

triggers = {
  payment_received = function (payment)
    if payment.extra ~= nil then
      db.ticket.update(payment.extra.ticket, { is_paid = true })
    end
  end
}
