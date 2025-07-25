# budva43 परियोजना के बारे में

## विवरण

**budva43** एक बुद्धिमान Telegram संदेश स्वचालित अग्रेषण प्रणाली है, जो Go भाषा में लिखी गई है। यह परियोजना एक enterprise-स्तरीय समाधान है जो UNIX-way सिद्धांतों और स्वच्छ आर्किटेक्चर को लागू करती है, विभिन्न चैनलों और समूहों के संदेशों से विषयगत सारांश बनाने के लिए।

## मुख्य कार्यक्षमता

### स्वचालित संदेश अग्रेषण
- **अग्रेषण (Forward)** — मूल लेखकत्व को संरक्षित करते हुए संदेश भेजना
- **प्रति भेजना (Send Copy)** — मूल स्रोत का संकेत दिए बिना संदेश प्रतियां बनाना
- **मीडिया एल्बम समर्थन** — समूहित छवियों और फ़ाइलों की प्रसंस्करण

### फ़िल्टरिंग सिस्टम
- **बहिष्करण फ़िल्टर** — अवांछित सामग्री को ब्लॉक करने के लिए नियमित अभिव्यक्तियां
- **समावेशन फ़िल्टर** — केवल प्रासंगिक संदेशों को पारित करने के नियम
- **सबस्ट्रिंग फ़िल्टर** — नियमित अभिव्यक्ति समूहों का उपयोग करके सटीक सबस्ट्रिंग फ़िल्टरिंग
- **स्वचालित उत्तर** — कीबोर्ड के साथ संदेशों के लिए स्वचालित प्रतिक्रियाएं
- **विशेष चैट** — फ़िल्टर किए गए संदेशों को check/अन्य चैनलों में स्वचालित भेजना

### सामग्री रूपांतरण
- **लिंक प्रतिस्थापन** — लक्षित चैट में अपने संदेशों के लिंक का स्वचालित प्रतिस्थापन
- **बाहरी लिंक हटाना** — बाहरी स्रोतों के लिंक की सफाई
- **पाठ खंड प्रतिस्थापन** — विभिन्न प्राप्तकर्ताओं के लिए अनुकूलन योग्य पाठ रूपांतरण
- **स्रोत हस्ताक्षर** — मूल संदेश स्रोत का संकेत
- **स्रोत लिंक जनरेशन** — मूल संदेशों के लिंक का स्वचालित जोड़ना

### संदेश जीवनचक्र प्रबंधन
- **एक बार कॉपी (Copy Once)** — संपादन पर समकालीकरण के बिना एकल भेजना
- **अमिट (Indelible)** — मूल के हटाए जाने पर संदेशों को हटाने से सुरक्षा
- **संपादन समकालीकरण** — मूल के बदलने पर कॉपी किए गए संदेशों का स्वचालित अद्यतन
- **हटाने का समकालीकरण** — स्रोत संदेश हटाए जाने पर प्रतियों को हटाना

### अतिरिक्त सुविधाएं
- **दर सीमा** — ब्लॉक को रोकने के लिए भेजने की गति नियंत्रण
- **सिस्टम संदेश प्रसंस्करण** — सेवा अधिसूचनाओं का स्वचालित हटाना

## आर्किटेक्चर

### माइक्रो सेवा संरचना
परियोजना दो मुख्य सेवाओं में विभाजित है:

#### इंजन (cmd/engine)
- **उद्देश्य**: संदेश अग्रेषण निष्पादन
- **प्रतिबंध**: आउटगोइंग चैट में नए संदेश भेजना निषिद्ध
- **घटक**: Telegram अद्यतन हैंडलर, अग्रेषण और फ़िल्टरिंग सेवाएं

#### फेसाड (cmd/facade)  
- **उद्देश्य**: API प्रदान करना (GraphQL, gRPC, REST)
- **क्षमताएं**: संदेश भेजने के कार्यों तक पूर्ण पहुंच
- **इंटरफेस**: वेब इंटरफेस, gRPC API, टर्मिनल इंटरफेस

### स्तरित आर्किटेक्चर

```
Transport Layer    → HTTP, gRPC, Terminal, Telegram Bot API
Service Layer      → व्यावसायिक तर्क, अग्रेषण नियम प्रसंस्करण
Repository Layer   → TDLib, Storage, Queue
Domain Layer       → डेटा मॉडल, अग्रेषण नियम
```

### डिज़ाइन पैटर्न
- **स्वच्छ आर्किटेक्चर** — जिम्मेदारी स्तरों का स्पष्ट विभाजन
- **निर्भरता इंजेक्शन** — कस्टम निर्भरता इंजेक्शन सिस्टम
- **रिपॉजिटरी पैटर्न** — डेटा पहुंच अमूर्तता
- **ऑब्जर्वर पैटर्न** — Telegram अद्यतन हैंडलिंग

## तकनीकी स्टैक

### मुख्य तकनीकें
- **Go 1.24** — generics समर्थन के साथ मुख्य विकास भाषा
- **TDLib** — क्लाइंट एप्लिकेशन के लिए आधिकारिक Telegram लाइब्रेरी
- **BadgerDB** — एम्बेडेड NoSQL डेटाबेस

### API और परिवहन
- **gRPC** — एकीकरण के लिए उच्च प्रदर्शन API
- **GraphQL** — वेब क्लाइंट के लिए लचीला API  
- **REST** — क्लासिक HTTP API
- **Telegram Client API** — Telegram के साथ प्रत्यक्ष बातचीत
- **टर्मिनल इंटरफेस** — इंटरैक्टिव टर्मिनल इंटरफेस

### विकास और परीक्षण
- **Docker और DevContainers** — विकास कंटेनराइज़ेशन
- **Testcontainers** — एकीकरण परीक्षण (Redis कनेक्शन परीक्षण सहित)
- **Mockery** — मॉक ऑब्जेक्ट जनरेशन
- **Godog (BDD)** — व्यवहार-संचालित परीक्षण
- **GitHub Actions CI** — स्वचालित निरंतर एकीकरण

### मॉनिटरिंग और अवलोकनीयता
- **संरचित लॉगिंग** — slog के साथ संरचित लॉग रिकॉर्डिंग
- **Grafana + Loki** — केंद्रीकृत लॉग और मॉनिटरिंग
- **pplog** — विकास के लिए मानव-पठनीय JSON लॉगिंग
- **spylog** — परीक्षणों में लॉग अवरोधन

### निर्माण और विकास उपकरण
- **Task** — कार्य स्वचालन के लिए Make विकल्प
- **golangci-lint** — व्यापक कोड गुणवत्ता जांच
- **कस्टम लिंटर** — "error-log-or-return" और "unused-interface-methods"
- **protobuf** — gRPC इंटरफेस जनरेशन
- **jq** — फ़िल्टरिंग के साथ वास्तविक समय लॉग देखना
- **EditorConfig** — संपादक सेटिंग्स स्थिरता

## विकास सिद्धांत

### आर्किटेक्चरल सिद्धांत
- **SOLID** — सभी पांच OOP सिद्धांतों का अनुप्रयोग
- **DRY** — कट्टरता के बिना कोड दुराव से बचना
- **KISS** — जटिल समाधानों पर सरल समाधानों को प्राथमिकता
- **YAGNI** — केवल आवश्यक कार्यक्षमता का कार्यान्वयन

### Go-विशिष्ट दृष्टिकोण
- **CSP (संचारी अनुक्रमिक प्रक्रियाएं)** — म्यूटेक्स के बजाय चैनलों का उपयोग
- **इंटरफेस पृथक्करण** — उपभोग करने वाले मॉड्यूल में स्थानीय इंटरफेस
- **इंटरफेस स्वीकार करें, संरचनाएं वापस करें** — इंटरफेस के साथ मुहावरेदार काम
- **प्रारंभिक वापसी** — कोड नेस्टिंग को कम करना
- **तालिका-संचालित परीक्षण** — संरचित परीक्षण

### त्रुटि हैंडलिंग सम्मेलन
- **संरचित त्रुटियां** — स्वचालित कॉल स्टैक के साथ संरचित त्रुटियां
- **लॉग या वापसी** — या तो त्रुटि लॉग करें या इसे ऊपर वापस करें
- **न्यूनतम रैपिंग** — केवल संदर्भ जोड़ते समय त्रुटियों को रैप करना

## कॉन्फ़िगरेशन

### सेटिंग्स पदानुक्रम
```
defaultConfig() → config.yml → .env
```

### कॉन्फ़िगरेशन प्रकार
- **स्थिर कॉन्फ़िगरेशन** — बुनियादी एप्लिकेशन सेटिंग्स
- **गतिशील कॉन्फ़िगरेशन** — हॉट-रीलोड के साथ अग्रेषण नियम
- **गुप्त डेटा** — पर्यावरण चर के माध्यम से API कुंजी और टोकन

### अग्रेषण कॉन्फ़िगरेशन उदाहरण
```yaml
forward_rules:
  rule1:
    from: 1001234567890
    to: [1009876543210, 1001111111111]
    send_copy: true
    exclude: "EXCLUDE|spam"
    include: "IMPORTANT|urgent"
    copy_once: false
    indelible: true
```

## परीक्षण

### बहु-स्तरीय परीक्षण
- **यूनिट परीक्षण** — मॉक के साथ अलग घटक परीक्षण
- **एकीकरण परीक्षण** — घटक बातचीत परीक्षण
- **E2E परीक्षण** — gRPC API के माध्यम से पूर्ण उपयोगकर्ता परिदृश्य
- **BDD परीक्षण** — प्राकृतिक भाषा में व्यवहार विवरण
- **स्नैपशॉट परीक्षण** — संदर्भ स्नैपशॉट के साथ परीक्षण

### विशेष तकनीकें
- **सिंक परीक्षण** — समय और समरूपता परीक्षण
- **कॉल-संचालित परीक्षण** — तैयारी कार्यों के साथ तालिका परीक्षण
- **स्पाई लॉगिंग** — परीक्षणों में लॉग अवरोधन और सत्यापन

### परीक्षण कवरेज
- **Codecov.io एकीकरण** — स्वचालित कोड कवरेज ट्रैकिंग
- **एकीकरण परीक्षण कवरेज** — विशेष उपकरण
- **कार्यात्मक कवरेज** — सभी मुख्य उपयोग परिदृश्य
- **तकनीकी कवरेज** — आंतरिक कार्य और edge cases
- **BDD परिदृश्य** — उपयोगकर्ता कहानियां और व्यावसायिक नियम

## तैनाती और संचालन

### लॉन्च विकल्प
- **स्थानीय विकास** — होस्ट मशीन पर TDLib की प्रत्यक्ष स्थापना
- **DevContainer** — पूर्णतः पृथक विकास वातावरण
- **उत्पादन** — कंटेनराइज़्ड तैनाती

### मॉनिटरिंग और डिबगिंग
- **संरचित लॉग** — मशीन प्रसंस्करण के लिए JSON लॉग
- **मानव-पठनीय लॉग** — विकास के लिए pplog
- **स्वास्थ्य जांच** — सेवा स्थिति जांच
- **graceful shutdown** — कार्य की सही समाप्ति

### एकीकरण
- **Telegram क्लाइंट** — प्राधिकरण के साथ पूर्ण-सुविधा क्लाइंट
- **बाहरी API** — GraphQL, gRPC, REST के माध्यम से एकीकरण
- **संदेश कतारें** — असिंक्रोनस संदेश प्रसंस्करण

## विकास में योगदान

### परियोजना दर्शन
Budva43 को "तकनीकों को लागू करने के लिए मेरा सबसे अच्छा शिक्षण परियोजना — MVP से Enterprise स्तर तक" के रूप में स्थित किया गया है। परियोजना आधुनिक Go विकास दृष्टिकोणों का प्रदर्शन करती है, जिसमें भाषा की नवीनतम सुविधाएं और उद्योग की सर्वोत्तम प्रथाएं शामिल हैं।

### अनूठी विशेषताएं
- **प्रयोगात्मक दृष्टिकोण** — अत्याधुनिक Go क्षमताओं का उपयोग
- **व्यापक परीक्षण** — परीक्षण तकनीकों का पूर्ण स्पेक्ट्रम
- **उत्पादन-तैयार गुणवत्ता** — औद्योगिक उपयोग के लिए तैयारी
- **शैक्षिक चरित्र** — समृद्ध दस्तावेज़ीकरण और उदाहरण

परियोजना सक्रिय रूप से विकसित हो रही है और enterprise विकास में आधुनिक Go क्षमताओं के प्रदर्शन के रूप में कार्य करती है। 