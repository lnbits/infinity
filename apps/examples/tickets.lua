title = "Tickets"

description = [[
Create buckets and then allow your users to create tickets on each of these buckets.

Only the tickets paid will be shown to you by default.
]]

models = {
  {
    name = 'bucket',
    display = 'Ticket Bucket',
    fields = {
      { name = 'description', display = 'Description', type = 'string' },
      { name = 'price', display = 'Price', type = 'msatoshi', required = true },
      { name = 'url', display = 'URL', type = 'url', computed = function (bucket)
        return bucket.key
      end }
    }
  },
  {
    name = 'ticket',
    display = 'Ticket',
    fields = {
      {
        name = 'bucket',
        display = 'Bucket',
        type = 'ref',
        ref = 'bucket',
        as = 'description',
      },
      { name = 'content', display = 'Content', type = 'string', required = true },
      { name = 'author', display = 'Author', type = 'string' },
      { name = 'is_paid', display = 'Paid', type = 'boolean' },
    },
    default_filters = {
      {'is_paid', '=', true}
    }
  },
}

actions = {
  createticket = {
    fields = {
      { name = 'bucket', type = 'string', required = true },
      { name = 'content', type = 'string', required = true },
      { name = 'author', type = 'string' },
    },
    handler = function (params)
      local key, err = db.ticket.add({
        bucket = params.bucket,
        content = params.content,
        author = params.author,
        is_paid = false,
      })
      if err then error(err) end

      local bucket, err = db.bucket.get(params.bucket)
      if err then error(err) end

      local payment, err = wallet.create_invoice({
        msatoshi = bucket.price,
        description = 'Ticket on bucket ' .. params.bucket,
        extra = { ticket = key }
      })
      if err then error(err) end

      return {
        bolt11 = payment.bolt11,
        ticket = key,
      }
    end
  }
}

triggers = {
  payment_received = function (payment)
    if payment.extra ~= nil then
      db.ticket.update(payment.extra.ticket, { is_paid = true })
      app.emit_event('ticket-paid', payment.extra.ticket)
    end
  end
}

files = {
  ['*'] = '/tickets.html'
}
