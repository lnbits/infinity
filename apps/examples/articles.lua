models = {
  {
    name = 'article',
    fields = {
      { name = 'url', required = true, type = 'url' },
      { name = 'name', required = true, type = 'string' },
      { name = 'author', required = true, type = 'string' },
      { name = 'votes', required = true, type = 'number' },
    },
    default_sort = 'votes desc'
  }
}

actions = {
  getarticles = {
    handler = function ()
      local articles, err = db.article.list()
      print(articles)
      print(err)
      return articles
    end
  },
  vote = {
    fields = {
      { name = 'article_key', required = true, type = 'string' },
      { name = 'votes', required = true, type = 'number' },
    },
    handler = function (params)
      local invoice = wallet.create_invoice({
        msatoshi = params.votes * 1000 * 10,
        description = params.votes .. ' votes on article ' .. params.article_key,
        extra = { article = params.article_key }
      })
      return invoice.bolt11
    end
  }
}

triggers = {
  payment_received = function (payment)
    if payment.extra and payment.extra.article then
      local article = db.article.get(payment.extra.article)
      article.votes = article.votes + payment.amount / 10 / 1000
      db.article.set(payment.extra.article, article)

      app.emit_event('vote-ack', payment.extra.article)
    end
  end
}

files = {
  ['*'] = '/articles.html'
}
