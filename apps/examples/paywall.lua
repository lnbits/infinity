models = {
  {
    name = 'item',
    fields = {
      { name = 'url', type = 'string', required = true },
      { name = 'description', type = 'string', required = true },
      { name = 'price', type = 'msatoshi', required = true },
      { name = 'paywall', type = 'url', computed = function (item) return item.key end },
    }
  }
}

actions = {
  getinfo = {
    fields = {
      { name = 'item', type = 'string', required = true }
    },
    handler = function (params)
      local item, err = db.item.get(params.item)
      if err then error(err) end

      return {
        description = item.description,
        price = item.price,
      }
    end,
  },
  requestpay = {
    fields = {
      { name = 'viewer_key', type = 'string', required = true },
      { name = 'item', type = 'string', required = true },
    },
    handler = function (params)
      local item, err = db.item.get(params.item)
      if err then error(err) end

      local payment, err = wallet.create_invoice({
        extra = { item = params.item, viewer_key = params.viewer_key },
        msatoshi = item.price,
        description = 'Unlock paywall for item ' .. params.item
      })
      if err then error(err) end

      return payment.bolt11
    end,
  }
}

triggers = {
  payment_received = function (payment)
    if payment.tag == app.id and payment.extra ~= nil then
      local item, err = db.item.get(payment.extra.item)
      if err then error(err) end

      local encrypted_url, err = utils.aes_encrypt(item.url, payment.extra.viewer_key)
      print(encrypted_url, err)

      app.emit_event('paywall-unlocked', { encrypted_url = encrypted_url })
    end
  end,
}

files = {
  ['*'] = 'paywall.html'
}
