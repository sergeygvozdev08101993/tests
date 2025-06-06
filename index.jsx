import { isArray, isEmpty, omitBy } from 'lodash';
import moment from 'moment/moment';
import React, { useCallback, useEffect, useState } from 'react';
import DateRangePicker from 'react-bootstrap-daterangepicker';
import { useMutation } from 'react-query';
import { useSearchParams } from 'react-router-dom';
import Select from 'react-select';
import { TextField } from 'src/components/base';
import InputStyles from 'src/components/base/textField/index.module.scss';
import { CallBackModalWindow } from 'src/components/callBackModalWindow';
import { GlobalStatesControl } from 'src/components/globalStates';
import { Header } from 'src/components/header';
import { Paginator, usePagination } from 'src/components/paginator';
import { useFetch } from 'src/hooks/useFetch';
import { localeSettings } from 'src/pages/const';
import { MailServerTable } from 'src/pages/email/mailserver/components/table';
import { getMailServerData, restartServerMail } from 'src/pages/email/mailserver/handlers';
import { statuses, results, sources } from 'src/pages/email/mailserver/models';
import { countItemsOnPage } from 'src/pages/generation/queue/models';
import { dateRange } from 'src/pages/wh-api/queue/models';
import { filterTableSelectorStyle } from 'src/styles/selectors/filterTable';
import { updatePageSelectorStyle } from 'src/styles/selectors/updatePage';

export const MailServerPage = () => {
  const [searchParams, setSearchParams] = useSearchParams();

  const initParams = Object.fromEntries([...searchParams]);

  const [itemsOnPage, setItemsOnPage] = useState({ value: 25, label: '25' });

  const [lengthPage, setLengthPage] = useState(initParams.length ? Number(initParams.length) : 25);

  const [modalIsOpen, setIsOpenModal] = useState(false);

  const [dataModal, setDataModal] = useState({});

  const triggerModal = active => {
    setIsOpenModal(active);
  };

  const [pageLoad, mailServerData, , { refresh }] = useFetch(
    () => getMailServerData(Object.fromEntries([...searchParams])),
    [window.location.search],
    {
      id: 'mail-server-data',
      enabled: !isEmpty([...searchParams]),
    }
  );

  const restartMail = useMutation(restartServerMail, {
    onSuccess: () => {
      refresh();

      triggerModal(false);
    },
    onError: err => {
      console.log('restartMail[err]: ', err);
    },
  });

  const [mailsFilter, setMailsFilter] = useState({
    filterId: '', // column 1
    filterFrom: '', // column 2
    filterTo: '', // column 3
    filterSubject: '', // column 4
    filterStatus: '', // column 5
    filterResult: '', // column 6
    filterSince: '', // column 8
    filterUntil: '', // column 8
    filterSource: '', // column 10
  });

  // в переменную сохраняются данные поля ввода для последующей проверки на изменение
  const [temporaryDataFilter, setTemporaryDataFilter] = useState('');

  const { firstContent, totalPages, page, setPage, pages } = usePagination({
    contentPerPage: lengthPage,
    count: mailServerData?.meta?.count,
    isLoading: pageLoad,
  });

  // getParams предназначена для получения параметров поиска.
  const getParams = (name, value, toFirstPage = false) => {
    const params = Object.fromEntries([...searchParams]);

    params.start = toFirstPage ? 0 : (page - 1) * lengthPage;
    params.length = lengthPage;

    if (name) {
      isArray(name)
        ? name.forEach((nameVal, index) => {
            params[nameVal] = value[index];
          })
        : (params[name] = value);
    } else {
      for (const key in mailsFilter) {
        if (mailsFilter[key] !== '') {
          params[key] = mailsFilter[key];
        }
      }
    }

    const omittedParams = omitBy(params, value => !value && value !== 0);

    setSearchParams(omittedParams);

    return params;
  };

  const handlerSetLengthPage = newValue => {
    setItemsOnPage(newValue);
    setLengthPage(newValue.value);
    getParams('length', newValue.value);
  };

  // Установка начальных данных
  useEffect(() => {
    getParams(['filterSince', 'filterUntil'], ['', '']);
    setPage(page, mailServerData?.meta?.count);

    const urlParams = new URLSearchParams(window.location.search);

    switch (urlParams.get('length')) {
      case '10':
        setItemsOnPage({ value: 10, label: '10' });
        break;
      case '25':
        setItemsOnPage({ value: 25, label: '25' });
        break;
      case '50':
        setItemsOnPage({ value: 50, label: '50' });
        break;
      case '100':
        setItemsOnPage({ value: 100, label: '100' });
        break;
      default:
    }

    //   eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    getParams();
    setPage(page, mailServerData?.meta?.count);
    setMailsFilter({ ...mailsFilter, ...initParams });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const onPageChange = (page, count) => {
    setPage(page, count);
    getParams('start', (page - 1) * lengthPage);
  };

  // quickSearch применяет выбранный параметр к фильтру и последующему поиску.
  const quickSearch = (name, value, toFirstPage = false) => {
    if (isArray(name)) {
      const newObject = {};

      // eslint-disable-next-line no-return-assign
      name.forEach((nameVal, index) => (newObject[nameVal] = value[index]));

      setMailsFilter({ ...mailsFilter, ...newObject });
    } else {
      setMailsFilter({ ...mailsFilter, [name]: value });
    }

    getParams(name, value, toFirstPage);

    if (toFirstPage) {
      setPage(1, mailServerData?.meta?.count);
    }
  };

  // handlerChangeStatus устанавливает новое значение
  // для фильтра по статусу.
  const handlerChangeStatus = async newValue => {
    quickSearch('filterStatus', newValue.value);
  };

  // handlerChangeResult устанавливает новое значение
  // для фильтра по результату.
  const handlerChangeResult = async newValue => {
    quickSearch('filterResult', newValue.value);
  };

  // handlerChangeSource устанавливает новое значение
  // для фильтра по источнику.
  const handlerChangeSource = async newValue => {
    quickSearch('filterSource', newValue.value);
  };

  // getFilterValue возвращает текущее значение фильтра.
  const getFilterValue = (value, options) => options?.find(option => option.value === value);

  // listenerEnterSearch ожидает нажатия на Enter для <input>.
  const listenerEnterSearch = event => {
    if (event.code === 'Enter' || event.code === 'NumpadEnter') {
      event.preventDefault();
      document.getElementById(event.target.id).blur();
    }
  };

  // searchByFilters содержит триггер поиска по фильтрам, обновляет данные писем.
  const searchByFilters = useCallback(
    event => {
      const clearMark = document.querySelector(`#clearMark-${event.target.name}`);

      if (clearMark !== null) {
        document.getElementById(event.target.id).removeEventListener('keydown', listenerEnterSearch);
        if (temporaryDataFilter !== event.target.value && !clearMark.isSameNode(event.relatedTarget)) {
          setPage(1, mailServerData?.meta?.count);
          getParams();
        }
      } else {
        document.getElementById(event.target.id).removeEventListener('keydown', listenerEnterSearch);
        quickSearch(event.target.name, '');
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [mailServerData?.meta?.count, temporaryDataFilter, getParams]
  );

  // onChange обновляет данные фильтров при вводе в <input>.
  const onChange = useCallback(
    ({ target: { name, value } }) => {
      setMailsFilter({ ...mailsFilter, [name]: value });
    },
    [mailsFilter]
  );

  // addListenerEnterSearch добавляет listenerEnterSearch к <input>.
  const addListenerEnterSearch = useCallback(event => {
    setTemporaryDataFilter(event.target.value);
    document.getElementById(event.target.id).addEventListener('keydown', listenerEnterSearch);
  }, []);

  const handleApply = (_, picker) => {
    const withMinutes = picker.chosenLabel === 'За 5 минут' || picker.chosenLabel === 'За 30 минут';

    const startDate = withMinutes
      ? moment(picker.startDate.i).format('YYYY-MM-DD HH:mm:ss')
      : picker.startDate.format('YYYY-MM-DD HH:mm:ss');

    const endDate = withMinutes ? moment(picker.endDate.i).format('YYYY-MM-DD HH:mm:ss') : picker.endDate.format('YYYY-MM-DD HH:mm:ss');

    let date = '';

    let dtSince = '';

    let dtUntil = '';

    if (startDate !== 'Invalid date' && !startDate.includes('2000-01-01')) {
      dtSince = startDate;
      if (startDate === endDate) {
        date = startDate;
      } else {
        dtUntil = endDate;
        date = `${startDate} - ${endDate}`;
      }
    }

    picker.element.val(date);
    quickSearch(['filterSince', 'filterUntil'], [dtSince, dtUntil]);
  };

  const handleCancel = (_, picker) => {
    picker.element.val('');
  };

  const triggerRestartMail = mail => {
    setDataModal({
      ...dataModal,
      title: 'Перезапустить письмо?',
      body: `Будет создана копия текущего письма id=${mail.id}.`,
      confirm: () => {
        restartMail.mutate({ id: mail.id });
      },
    });

    triggerModal(true);
  };

  const tableHeader = [
    {
      name: 'ID',
      filter: (
        <div className="ts-wrap-input" data-validate="Enter password">
          <TextField
            id="filterId"
            name="filterId"
            placeholder="&#xe82a;"
            title="Фильтр по идентификатору письма"
            value={mailsFilter.filterId}
            onBlur={searchByFilters}
            onChange={onChange}
            onClear={() => quickSearch('filterId', '')}
            onFocus={addListenerEnterSearch}
          />
        </div>
      ),
    },
    {
      name: 'От кого',
      filter: (
        <div className="ts-wrap-input" data-validate="Enter password">
          <TextField
            id="filterFrom"
            name="filterFrom"
            placeholder="&#xe818;"
            title="Фильтр по адресу отправителя письма"
            value={mailsFilter.filterFrom}
            onBlur={searchByFilters}
            onChange={onChange}
            onClear={() => quickSearch('filterFrom', '')}
            onFocus={addListenerEnterSearch}
          />
        </div>
      ),
    },
    {
      name: 'Кому',
      filter: (
        <div className="ts-wrap-input" data-validate="Enter password">
          <TextField
            id="filterTo"
            name="filterTo"
            placeholder="&#xe818;"
            title="Фильтр по адресу получателя письма"
            value={mailsFilter.filterTo}
            onBlur={searchByFilters}
            onChange={onChange}
            onClear={() => quickSearch('filterTo', '')}
            onFocus={addListenerEnterSearch}
          />
        </div>
      ),
    },
    {
      name: 'Тема',
      filter: (
        <div className="ts-wrap-input" data-validate="Enter password">
          <TextField
            id="filterSubject"
            name="filterSubject"
            placeholder="&#xe818;"
            title="Фильтр по теме письма"
            value={mailsFilter.filterSubject}
            onBlur={searchByFilters}
            onChange={onChange}
            onClear={() => quickSearch('filterSubject', '')}
            onFocus={addListenerEnterSearch}
          />
        </div>
      ),
    },
    {
      name: 'Статус',
      filter: (
        <Select
          options={statuses}
          styles={filterTableSelectorStyle}
          value={getFilterValue(mailsFilter.filterStatus, statuses)}
          onChange={handlerChangeStatus}
        />
      ),
    },
    {
      name: 'Результат',
      filter: (
        <Select
          options={results}
          styles={filterTableSelectorStyle}
          value={getFilterValue(mailsFilter.filterResult, results)}
          onChange={handlerChangeResult}
        />
      ),
    },
    {
      name: 'Eml',
    },
    {
      name: 'Принято',
      filter: (
        <div className="ts-wrap-input">
          <DateRangePicker
            id="filterPutTime"
            initialSettings={{
              startDate: mailsFilter.filterSince ? moment(mailsFilter.filterSince).toDate() : moment().toDate(),
              endDate: mailsFilter.filterUntil ? moment(mailsFilter.filterUntil).toDate() : moment().toDate(),
              autoUpdateInput: false,
              ranges: dateRange,
              locale: localeSettings,
              timePicker: true,
            }}
            onApply={handleApply}
            onCancel={handleCancel}
          >
            <input
              autoComplete="off"
              className={InputStyles['ts-input']}
              defaultValue=""
              id="reportrange"
              name="filterTime"
              placeholder="&#xe87f;"
              type="search"
            />
          </DateRangePicker>
        </div>
      ),
    },
    {
      name: 'Обработано',
    },
    {
      name: 'Источник',
      filter: (
        <Select
          options={sources}
          styles={filterTableSelectorStyle}
          value={getFilterValue(mailsFilter.filterSource, sources)}
          onChange={handlerChangeSource}
        />
      ),
    },
  ];

  const updateDataWithButton = useCallback(() => !pageLoad && refresh(), [pageLoad, refresh]);

  return (
    <div className="generation-wrap">
      <div className="generation-header">
        <Header
          withoutTimer
          title={
            <>
              Почта
              <div className="navigation-arrow-right" />
              Почтовый сервер ({mailServerData?.meta?.count})
            </>
          }
          updateData={updateDataWithButton}
        >
          <GlobalStatesControl />
        </Header>

        <div className="navigation-page-header-bottom-wrap">
          <Paginator
            PgCount={mailServerData?.meta?.count}
            PgFirstContent={firstContent}
            PgPage={page}
            PgPageLoad={pageLoad}
            PgPages={pages}
            PgSetPage={onPageChange}
            PgTotalPages={totalPages}
          />

          <div className="navigation-page-header-cp-wrap">
            <div className="navigation-page-header-configs-input" title="Количество записей на странице">
              <Select
                defaultValue={countItemsOnPage[1]}
                options={countItemsOnPage}
                styles={updatePageSelectorStyle}
                value={itemsOnPage}
                onChange={handlerSetLengthPage}
              />
            </div>
          </div>
        </div>
      </div>

      <div className="mainTable-wrap">
        <MailServerTable
          headerList={tableHeader}
          pageLoad={pageLoad}
          quickSearch={quickSearch}
          rows={mailServerData?.body || []}
          triggerRestartMail={triggerRestartMail}
        />
      </div>

      <div className="generation-footer d-flex justify-content-end">
        <Paginator
          PgCount={mailServerData?.meta?.count}
          PgFirstContent={firstContent}
          PgPage={page}
          PgPageLoad={pageLoad}
          PgPages={pages}
          PgSetPage={onPageChange}
          PgTotalPages={totalPages}
        />
      </div>

      <CallBackModalWindow
        activeModal={modalIsOpen}
        closeModal={() => {
          triggerModal(false);
        }}
        dataModal={dataModal}
      />
    </div>
  );
};
